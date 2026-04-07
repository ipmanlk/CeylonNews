package fetcher

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"ipmanlk/cnapi/internal/model"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-shiori/dom"
	"github.com/markusmobius/go-trafilatura"
	"github.com/mmcdole/gofeed"
)

type Fetcher struct {
	httpClient    *HTTPClient
	browserClient *BrowserAPIClient
}

func NewFetcher(httpClient *HTTPClient, browserClient *BrowserAPIClient) *Fetcher {
	return &Fetcher{
		httpClient:    httpClient,
		browserClient: browserClient,
	}
}

func (f *Fetcher) FetchHTML(ctx context.Context, url string, useBrowser bool) ([]byte, error) {
	if useBrowser {
		return f.browserClient.FetchHTML(ctx, url)
	}
	return f.httpClient.FetchHTML(ctx, url)
}

func (f *Fetcher) FetchHTMLDoc(ctx context.Context, url string, useBrowser bool) (*goquery.Document, error) {
	if useBrowser {
		return f.browserClient.FetchHTMLDoc(ctx, url)
	}
	return f.httpClient.FetchHTMLDoc(ctx, url)
}

func (f *Fetcher) ExtractArticle(ctx context.Context, url string, useBrowser bool) (*trafilatura.ExtractResult, error) {
	html, err := f.FetchHTML(ctx, url, useBrowser)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch HTML from %s: %w", url, err)
	}

	opts := trafilatura.Options{
		IncludeLinks:    true,
		IncludeImages:   true,
		ExcludeComments: true,
		EnableFallback:  true,
		Deduplicate:     true,
	}

	result, err := trafilatura.Extract(bytes.NewReader(html), opts)
	if err != nil {
		return nil, fmt.Errorf("failed to extract content from %s: %w", url, err)
	}

	return result, nil
}

func (f *Fetcher) FetchRSS(ctx context.Context, url string, maxItems int) ([]*gofeed.Item, error) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURLWithContext(url, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSS feed: %w", err)
	}

	return deduplicateRSSItems(feed.Items, maxItems, url), nil
}

func (f *Fetcher) FetchRSSWithBrowser(ctx context.Context, url string, maxItems int) ([]*gofeed.Item, error) {
	html, err := f.browserClient.FetchHTML(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch RSS feed with browser: %w", err)
	}

	fp := gofeed.NewParser()
	feed, err := fp.ParseString(string(html))
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSS feed from browser content: %w", err)
	}

	return deduplicateRSSItems(feed.Items, maxItems, url), nil
}

func deduplicateRSSItems(items []*gofeed.Item, maxItems int, feedURL string) []*gofeed.Item {
	if maxItems > 0 && len(items) > maxItems {
		items = items[:maxItems]
	}

	seen := make(map[string]struct{}, len(items))
	result := make([]*gofeed.Item, 0, len(items))

	for _, item := range items {
		if item.Link == "" {
			slog.Warn("skipping item with empty link", "feed_url", feedURL, "item_title", item.Title)
			continue
		}
		if _, exists := seen[item.Link]; exists {
			slog.Debug("skipping duplicate item", "feed_url", feedURL, "item_link", item.Link)
			continue
		}
		seen[item.Link] = struct{}{}
		result = append(result, item)
	}

	return result
}

func (f *Fetcher) ExtractArticleFromRSSItem(ctx context.Context, item *gofeed.Item, useBrowser bool) (model.ScrapedArticle, error) {
	result, err := f.ExtractArticle(ctx, item.Link, useBrowser)
	if err != nil {
		return model.ScrapedArticle{}, fmt.Errorf("failed to extract article: %w", err)
	}

	return f.buildArticleFromRSSItem(item, result), nil
}

func (f *Fetcher) ExtractArticleFromRSSItemWithSelector(ctx context.Context, item *gofeed.Item, contentSelector string, useBrowser bool) (model.ScrapedArticle, error) {
	result, err := f.ExtractArticleWithSelector(ctx, item.Link, contentSelector, useBrowser)
	if err != nil {
		return model.ScrapedArticle{}, fmt.Errorf("failed to extract article: %w", err)
	}

	return f.buildArticleFromRSSItem(item, result), nil
}

func (f *Fetcher) buildArticleFromRSSItem(item *gofeed.Item, result *trafilatura.ExtractResult) model.ScrapedArticle {
	doc := trafilatura.CreateReadableDocument(result)
	htmlContent := dom.OuterHTML(doc)

	return model.ScrapedArticle{
		Title:       item.Title,
		URL:         item.Link,
		ContentText: result.ContentText,
		ContentHTML: htmlContent,
		ImageURL:    f.getImageURL(item),
		Categories:  item.Categories,
		PublishedAt: f.getPublishedAt(item),
	}
}

func (f *Fetcher) ExtractArticleWithSelector(ctx context.Context, url string, contentSelector string, useBrowser bool) (*trafilatura.ExtractResult, error) {
	doc, err := f.FetchHTMLDoc(ctx, url, useBrowser)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch HTML doc from %s: %w", url, err)
	}

	contentNode, err := f.extractNodeWithSelector(doc, contentSelector)
	if err != nil {
		return nil, fmt.Errorf("failed to extract content node from %s: %w", url, err)
	}

	opts := trafilatura.Options{
		IncludeLinks:    true,
		IncludeImages:   true,
		ExcludeComments: true,
		EnableFallback:  true,
		Deduplicate:     true,
	}

	result, err := trafilatura.ExtractDocument(contentNode.Get(0), opts)
	if err != nil {
		return nil, fmt.Errorf("failed to extract content from %s: %w", url, err)
	}

	return result, nil
}

func (f *Fetcher) ExtractLinks(doc *goquery.Document, selector, urlPrefix string) []string {
	var links []string
	doc.Find(selector).Each(func(i int, selection *goquery.Selection) {
		href, exists := selection.Attr("href")
		if exists && href != "" && strings.HasPrefix(href, urlPrefix) {
			links = append(links, href)
		}
	})
	return links
}

func (f *Fetcher) CreateScrapedArticle(sourceName string, result *trafilatura.ExtractResult, url string, imageURL *string, publishedAt time.Time) model.ScrapedArticle {
	doc := trafilatura.CreateReadableDocument(result)
	htmlContent := dom.OuterHTML(doc)

	return model.ScrapedArticle{
		SourceName:  sourceName,
		Title:       result.Metadata.Title,
		URL:         url,
		ContentText: result.ContentText,
		ContentHTML: htmlContent,
		ImageURL:    imageURL,
		PublishedAt: publishedAt,
	}
}

func (f *Fetcher) getImageURL(item *gofeed.Item) *string {
	if item.Image != nil {
		return &item.Image.URL
	}

	if url := f.getFirstImageFromHTML([]byte(item.Content)); url != nil {
		return url
	}

	if url := f.getFirstImageFromHTML([]byte(item.Description)); url != nil {
		return url
	}

	imageTypes := map[string]struct{}{
		"image/jpeg": {},
		"image/png":  {},
		"image/webp": {},
		"image/jpg":  {},
	}

	for _, enclosure := range item.Enclosures {
		if _, ok := imageTypes[enclosure.Type]; ok {
			return &enclosure.URL
		}
	}

	return nil
}

func (f *Fetcher) getPublishedAt(item *gofeed.Item) time.Time {
	if item.PublishedParsed != nil {
		return *item.PublishedParsed
	}
	if item.UpdatedParsed != nil {
		return *item.UpdatedParsed
	}
	return time.Now()
}

func (f *Fetcher) getFirstImageFromHTML(html []byte) *string {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		return nil
	}
	src := doc.Find("img").First().AttrOr("src", "")
	if src == "" {
		return nil
	}
	return &src
}

func (f *Fetcher) extractNodeWithSelector(doc *goquery.Document, selector string) (*goquery.Selection, error) {
	selection := doc.Find(selector)
	if selection.Length() == 0 {
		return nil, fmt.Errorf("no elements found with selector: %s", selector)
	}
	return selection, nil
}
