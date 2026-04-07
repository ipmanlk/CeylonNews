package fetcher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// BrowserAPIClient fetches pages via an external headless browser API service.
type BrowserAPIClient struct {
	apiURL      string
	client      *http.Client
	waitTimeSec int
}

type browserAPIResponse struct {
	Success bool   `json:"success"`
	HTML    string `json:"html"`
	URL     string `json:"url"`
	Error   string `json:"error,omitempty"`
}

func NewBrowserAPIClient(apiURL string, timeout time.Duration, waitTimeSec int) *BrowserAPIClient {
	return &BrowserAPIClient{
		apiURL: apiURL,
		client: &http.Client{
			Timeout: timeout,
		},
		waitTimeSec: waitTimeSec,
	}
}

func (b *BrowserAPIClient) FetchHTML(ctx context.Context, targetURL string) ([]byte, error) {
	params := url.Values{}
	params.Set("url", targetURL)
	params.Set("wait_time", fmt.Sprintf("%d", b.waitTimeSec))
	apiURL := b.apiURL + "/scrape?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for browser API: %w", err)
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call browser API for %s: %w", targetURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("browser API returned non-200 status code (%d) for %s", resp.StatusCode, targetURL)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read browser API response: %w", err)
	}

	var apiResp browserAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse browser API response: %w", err)
	}

	if !apiResp.Success {
		return nil, fmt.Errorf("browser API failed for %s: %s", targetURL, apiResp.Error)
	}

	if apiResp.HTML == "" {
		return nil, fmt.Errorf("empty HTML response from browser API for %s", targetURL)
	}

	return []byte(apiResp.HTML), nil
}

func (b *BrowserAPIClient) FetchHTMLDoc(ctx context.Context, targetURL string) (*goquery.Document, error) {
	html, err := b.FetchHTML(ctx, targetURL)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML from browser API for %s: %w", targetURL, err)
	}

	return doc, nil
}
