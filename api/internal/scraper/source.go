package scraper

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"ipmanlk/cnapi/internal/fetcher"
	"ipmanlk/cnapi/internal/model"

	"github.com/mmcdole/gofeed"
)

type Source struct {
	config  Config
	fetcher *fetcher.Fetcher
}

func NewSource(cfg Config, f *fetcher.Fetcher) *Source {
	return &Source{config: cfg, fetcher: f}
}

func (s *Source) Name() string { return s.config.Name }

func (s *Source) Languages() []model.Language {
	langs := make([]model.Language, 0, len(s.config.Languages))
	for _, lc := range s.config.Languages {
		langs = append(langs, model.Language(lc.Language))
	}
	return langs
}

func (s *Source) Scrape(ctx context.Context, language model.Language) ([]model.ScrapedArticle, error) {
	for _, lc := range s.config.Languages {
		if model.Language(lc.Language) == language {
			return s.scrapeLanguage(ctx, lc)
		}
	}
	return nil, nil
}

func (s *Source) UsesBrowser(language model.Language) bool {
	for _, lc := range s.config.Languages {
		if model.Language(lc.Language) == language {
			return lc.Listing.Browser || lc.Article.NeedsBrowser()
		}
	}
	return false
}

func (s *Source) scrapeLanguage(ctx context.Context, lc LanguageConfig) ([]model.ScrapedArticle, error) {
	switch lc.Listing.Type {
	case "rss":
		return s.scrapeRSSListing(ctx, lc)
	case "html":
		return s.scrapeHTMLListing(ctx, lc)
	default:
		return nil, fmt.Errorf("unknown listing type %q for source %q", lc.Listing.Type, s.Name())
	}
}

func (s *Source) scrapeRSSListing(ctx context.Context, lc LanguageConfig) ([]model.ScrapedArticle, error) {
	var items []*gofeed.Item
	var err error

	if lc.Listing.Browser {
		items, err = s.fetcher.FetchRSSWithBrowser(ctx, lc.Listing.URL, lc.MaxItems)
	} else {
		items, err = s.fetcher.FetchRSS(ctx, lc.Listing.URL, lc.MaxItems)
	}
	if err != nil {
		return nil, err
	}

	return s.extractFromRSSItems(ctx, lc, items)
}

func (s *Source) scrapeHTMLListing(ctx context.Context, lc LanguageConfig) ([]model.ScrapedArticle, error) {
	doc, err := s.fetcher.FetchHTMLDoc(ctx, lc.Listing.URL, lc.Listing.Browser)
	if err != nil {
		return nil, err
	}

	var links []string
	for _, sel := range lc.Listing.Selectors {
		links = append(links, s.fetcher.ExtractLinks(doc, sel, lc.Listing.URLPrefix)...)
	}

	links = lc.Listing.ResolveLinks(links)

	if lc.MaxItems > 0 && len(links) > lc.MaxItems {
		links = links[:lc.MaxItems]
	}

	return s.extractFromLinks(ctx, lc, links)
}

func (s *Source) extractFromRSSItems(ctx context.Context, lc LanguageConfig, items []*gofeed.Item) ([]model.ScrapedArticle, error) {
	articles := make([]model.ScrapedArticle, 0, len(items))
	seen := make(map[string]bool)

	for _, item := range items {
		if seen[item.Link] {
			continue
		}
		seen[item.Link] = true

		if s.shouldSkip(item.Title) {
			slog.Debug("skipping article by title filter", "source", s.Name(), "url", item.Link)
			continue
		}

		var article model.ScrapedArticle
		var err error

		if lc.Article.Selector != "" {
			article, err = s.fetcher.ExtractArticleFromRSSItemWithSelector(ctx, item, lc.Article.Selector, lc.Article.Browser)
		} else {
			article, err = s.fetcher.ExtractArticleFromRSSItem(ctx, item, lc.Article.Browser)
		}

		if err != nil {
			slog.Warn("failed to extract article", "source", s.Name(), "url", item.Link, "error", err)
			continue
		}

		if article.Title == "" || article.ContentText == "" || article.ContentHTML == "" {
			slog.Debug("skipping article with missing content", "source", s.Name(), "url", item.Link)
			continue
		}

		article.Title = s.applyTitleReplace(article.Title)
		article.SourceName = s.Name()
		article.Language = model.Language(lc.Language)
		articles = append(articles, article)
	}

	return articles, nil
}

func (s *Source) extractFromLinks(ctx context.Context, lc LanguageConfig, links []string) ([]model.ScrapedArticle, error) {
	articles := make([]model.ScrapedArticle, 0, len(links))
	seen := make(map[string]bool)

	for _, link := range links {
		if seen[link] {
			continue
		}
		seen[link] = true

		result, err := s.fetcher.ExtractArticle(ctx, link, lc.Article.Browser)
		if err != nil {
			slog.Warn("failed to extract article", "source", s.Name(), "url", link, "error", err)
			continue
		}

		if result == nil || result.Metadata.Title == "" || result.ContentText == "" {
			slog.Debug("skipping article with missing content", "source", s.Name(), "url", link)
			continue
		}

		article := s.fetcher.CreateScrapedArticle(s.Name(), result, link, &result.Metadata.Image, result.Metadata.Date)
		if article.ContentText == "" || article.ContentHTML == "" {
			slog.Debug("skipping article with missing extracted content", "source", s.Name(), "url", link)
			continue
		}

		article.Language = model.Language(lc.Language)
		articles = append(articles, article)
	}

	return articles, nil
}

func (s *Source) shouldSkip(title string) bool {
	for _, rule := range s.config.TitleTransform.Skip {
		if rule.CaseSensitive {
			if strings.Contains(title, rule.Contains) {
				return true
			}
		} else {
			if strings.Contains(strings.ToLower(title), strings.ToLower(rule.Contains)) {
				return true
			}
		}
	}
	return false
}

func (s *Source) applyTitleReplace(title string) string {
	for _, rule := range s.config.TitleTransform.Replace {
		if rule.CaseSensitive {
			title = strings.ReplaceAll(title, rule.Pattern, rule.With)
		} else {
			title = replaceAllCaseInsensitive(title, rule.Pattern, rule.With)
		}
	}
	return title
}

func replaceAllCaseInsensitive(s, old, new string) string {
	if old == "" {
		return s
	}
	lowerS := strings.ToLower(s)
	lowerOld := strings.ToLower(old)
	var result strings.Builder
	start := 0
	for {
		idx := strings.Index(lowerS[start:], lowerOld)
		if idx == -1 {
			result.WriteString(s[start:])
			break
		}
		idx += start
		result.WriteString(s[start:idx])
		result.WriteString(new)
		start = idx + len(old)
	}
	return result.String()
}
