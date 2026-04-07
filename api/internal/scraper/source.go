package scraper

import (
	"context"
	"fmt"
	"log/slog"

	"ipmanlk/cnapi/internal/fetcher"
	"ipmanlk/cnapi/internal/model"

	"github.com/mmcdole/gofeed"
)

// Source represents a single news source with scraping capabilities
type Source struct {
	config  Config
	fetcher *fetcher.Fetcher
}

// NewSource creates a new Source instance
func NewSource(cfg Config, f *fetcher.Fetcher) *Source {
	return &Source{config: cfg, fetcher: f}
}

// Name returns the source name
func (s *Source) Name() string { return s.config.Name }

// Languages returns the list of supported languages
func (s *Source) Languages() []model.Language {
	langs := make([]model.Language, 0, len(s.config.Languages))
	for _, lc := range s.config.Languages {
		langs = append(langs, model.Language(lc.Language))
	}
	return langs
}

// Scrape scrapes articles for a given language
func (s *Source) Scrape(ctx context.Context, language model.Language) ([]model.ScrapedArticle, error) {
	for _, lc := range s.config.Languages {
		if model.Language(lc.Language) == language {
			return s.scrapeLanguage(ctx, lc)
		}
	}
	return nil, nil
}

// UsesBrowser checks if this source requires a browser for any operation
func (s *Source) UsesBrowser(language model.Language) bool {
	for _, lc := range s.config.Languages {
		if model.Language(lc.Language) == language {
			return lc.Discovery.Browser || lc.Extraction.Browser
		}
	}
	return false
}

// scrapeLanguage executes the full pipeline for one language
func (s *Source) scrapeLanguage(ctx context.Context, lc LanguageConfig) ([]model.ScrapedArticle, error) {
	switch lc.Discovery.Type {
	case "rss":
		return s.scrapeRSS(ctx, lc)
	case "html":
		return s.scrapeHTML(ctx, lc)
	default:
		return nil, fmt.Errorf("unknown discovery type %q for source %q", lc.Discovery.Type, s.Name())
	}
}

// scrapeRSS handles RSS-based discovery
func (s *Source) scrapeRSS(ctx context.Context, lc LanguageConfig) ([]model.ScrapedArticle, error) {
	var items []*gofeed.Item
	var err error

	if lc.Discovery.Browser {
		items, err = s.fetcher.FetchRSSWithBrowser(ctx, lc.Discovery.URL, lc.MaxItems)
	} else {
		items, err = s.fetcher.FetchRSS(ctx, lc.Discovery.URL, lc.MaxItems)
	}
	if err != nil {
		return nil, err
	}

	return s.extractFromRSSItems(ctx, lc, items)
}

// scrapeHTML handles HTML-based discovery
func (s *Source) scrapeHTML(ctx context.Context, lc LanguageConfig) ([]model.ScrapedArticle, error) {
	doc, err := s.fetcher.FetchHTMLDoc(ctx, lc.Discovery.URL, lc.Discovery.Browser)
	if err != nil {
		return nil, err
	}

	// Extract links from all selectors
	var links []string
	for _, sel := range lc.Discovery.HTML.LinkSelectors {
		links = append(links, s.fetcher.ExtractLinks(doc, sel, "")...)
	}

	// Apply URL transformation pipeline
	links = ApplyURLRules(links, lc.Discovery.HTML.URLRules)

	// Deduplicate links
	links = deduplicateStrings(links)

	// Limit to max items
	if lc.MaxItems > 0 && len(links) > lc.MaxItems {
		links = links[:lc.MaxItems]
	}

	// Early validation: extract titles from listing if selector provided
	if lc.Discovery.HTML.TitleSelector != "" {
		return s.extractWithEarlyValidation(ctx, lc, doc, links)
	}

	// Standard extraction without early validation
	return s.extractFromLinks(ctx, lc, links)
}

// extractWithEarlyValidation extracts titles from listing and validates before full fetch
func (s *Source) extractWithEarlyValidation(ctx context.Context, lc LanguageConfig, doc interface{}, links []string) ([]model.ScrapedArticle, error) {
	articles := make([]model.ScrapedArticle, 0, len(links))
	validationEngine := NewValidationEngine(lc.Validation)

	for _, link := range links {
		// Extract title from listing page
		title := s.fetcher.ExtractFieldFromDoc(doc, lc.Discovery.HTML.TitleSelector)

		// Early validation on title
		if title != "" {
			if keep, reason := validationEngine.ValidateTitle(title); !keep {
				slog.Debug("skipping article by early validation", "source", s.Name(), "url", link, "reason", reason)
				continue
			}
		}

		// Extract full article
		article, err := s.extractSingleArticle(ctx, lc, link)
		if err != nil {
			slog.Warn("failed to extract article", "source", s.Name(), "url", link, "error", err)
			continue
		}

		// Override title if we got it from listing
		if title != "" {
			article.Title = title
		}

		// Late validation (full article)
		if keep, reason := validationEngine.Validate(&article); !keep {
			slog.Debug("skipping article by validation", "source", s.Name(), "url", link, "reason", reason)
			continue
		}

		// Apply transformations
		transformationEngine := NewTransformationEngine(lc.Transformation)
		transformationEngine.Transform(&article)

		articles = append(articles, article)
	}

	return articles, nil
}

// extractFromRSSItems processes RSS feed items
func (s *Source) extractFromRSSItems(ctx context.Context, lc LanguageConfig, items []*gofeed.Item) ([]model.ScrapedArticle, error) {
	articles := make([]model.ScrapedArticle, 0, len(items))
	seen := make(map[string]bool)
	validationEngine := NewValidationEngine(lc.Validation)
	transformationEngine := NewTransformationEngine(lc.Transformation)

	for _, item := range items {
		if seen[item.Link] {
			continue
		}
		seen[item.Link] = true

		// Early validation on RSS title
		if item.Title != "" {
			if keep, reason := validationEngine.ValidateTitle(item.Title); !keep {
				slog.Debug("skipping article by early validation", "source", s.Name(), "url", item.Link, "reason", reason)
				continue
			}
		}

		// Extract article content
		var article model.ScrapedArticle
		var err error

		if lc.Extraction.Content.ScopeSelector != "" {
			article, err = s.fetcher.ExtractArticleFromRSSItemWithSelector(ctx, item, lc.Extraction.Content.ScopeSelector, lc.Extraction.Browser)
		} else {
			article, err = s.fetcher.ExtractArticleFromRSSItem(ctx, item, lc.Extraction.Browser)
		}

		if err != nil {
			slog.Warn("failed to extract article", "source", s.Name(), "url", item.Link, "error", err)
			continue
		}

		// Check for required fields
		if article.Title == "" || article.ContentText == "" || article.ContentHTML == "" {
			slog.Debug("skipping article with missing content", "source", s.Name(), "url", item.Link)
			continue
		}

		// Late validation
		if keep, reason := validationEngine.Validate(&article); !keep {
			slog.Debug("skipping article by validation", "source", s.Name(), "url", item.Link, "reason", reason)
			continue
		}

		// Apply transformations
		transformationEngine.Transform(&article)

		article.SourceName = s.Name()
		article.Language = model.Language(lc.Language)
		articles = append(articles, article)
	}

	return articles, nil
}

// extractFromLinks processes HTML links
func (s *Source) extractFromLinks(ctx context.Context, lc LanguageConfig, links []string) ([]model.ScrapedArticle, error) {
	articles := make([]model.ScrapedArticle, 0, len(links))
	seen := make(map[string]bool)
	validationEngine := NewValidationEngine(lc.Validation)
	transformationEngine := NewTransformationEngine(lc.Transformation)

	for _, link := range links {
		if seen[link] {
			continue
		}
		seen[link] = true

		// Extract article
		article, err := s.extractSingleArticle(ctx, lc, link)
		if err != nil {
			slog.Warn("failed to extract article", "source", s.Name(), "url", link, "error", err)
			continue
		}

		// Validation
		if keep, reason := validationEngine.Validate(&article); !keep {
			slog.Debug("skipping article by validation", "source", s.Name(), "url", link, "reason", reason)
			continue
		}

		// Apply transformations
		transformationEngine.Transform(&article)

		articles = append(articles, article)
	}

	return articles, nil
}

// extractSingleArticle extracts a single article from a URL
func (s *Source) extractSingleArticle(ctx context.Context, lc LanguageConfig, url string) (model.ScrapedArticle, error) {
	// Use content-specific extraction if selectors are provided
	if lc.Extraction.Content.TitleSelector != "" ||
		lc.Extraction.Content.BodySelector != "" ||
		lc.Extraction.Content.ScopeSelector != "" {
		contentConfig := fetcher.ContentConfig{
			ScopeSelector: lc.Extraction.Content.ScopeSelector,
			TitleSelector: lc.Extraction.Content.TitleSelector,
			BodySelector:  lc.Extraction.Content.BodySelector,
			ImageSelector: lc.Extraction.Content.ImageSelector,
			DateSelector:  lc.Extraction.Content.DateSelector,
		}
		return s.fetcher.ExtractArticleWithContentConfig(ctx, url, contentConfig, lc.Extraction.Browser)
	}

	// Use default trafilatura extraction
	result, err := s.fetcher.ExtractArticle(ctx, url, lc.Extraction.Browser)
	if err != nil {
		return model.ScrapedArticle{}, err
	}

	return s.fetcher.CreateScrapedArticle(s.Name(), result, url, &result.Metadata.Image, result.Metadata.Date), nil
}

// deduplicateStrings removes duplicates from a string slice
func deduplicateStrings(items []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(items))
	for _, item := range items {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	return result
}
