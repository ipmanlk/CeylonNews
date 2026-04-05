package fetcher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type BrowserAPIClient struct {
	apiURL   string
	client   *http.Client
	waitTime int
}

type BrowserAPIResponse struct {
	Success bool   `json:"success"`
	HTML    string `json:"html"`
	URL     string `json:"url"`
	Error   string `json:"error,omitempty"`
}

func NewBrowserAPIClient(apiURL string, timeout time.Duration, waitTime int) *BrowserAPIClient {
	return &BrowserAPIClient{
		apiURL: apiURL,
		client: &http.Client{
			Timeout: timeout,
		},
		waitTime: waitTime,
	}
}

func (b *BrowserAPIClient) FetchHTML(ctx context.Context, url string) ([]byte, error) {
	apiURL := fmt.Sprintf("%s/scrape?url=%s&wait_time=%d", b.apiURL, url, b.waitTime)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for browser API: %w", err)
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call browser API for %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("browser API returned non-200 status code (%d) for %s", resp.StatusCode, url)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read browser API response: %w", err)
	}

	var apiResp BrowserAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse browser API response: %w", err)
	}

	if !apiResp.Success {
		return nil, fmt.Errorf("browser API failed for %s: %s", url, apiResp.Error)
	}

	if apiResp.HTML == "" {
		return nil, fmt.Errorf("empty HTML response from browser API for %s", url)
	}

	return []byte(apiResp.HTML), nil
}

func (b *BrowserAPIClient) FetchHTMLDoc(ctx context.Context, url string) (*goquery.Document, error) {
	html, err := b.FetchHTML(ctx, url)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML from browser API for %s: %w", url, err)
	}

	return doc, nil
}
