package scraper

import (
	"context"
	"fmt"

	"ipmanlk/cnapi/internal/fetcher"
	"ipmanlk/cnapi/internal/model"
)

type SourceScraper interface {
	Name() string
	Languages() []model.Language
	Scrape(ctx context.Context, language model.Language) ([]model.ScrapedArticle, error)
	UsesBrowser(language model.Language) bool
}

type Registry struct {
	scrapers []SourceScraper
}

func NewRegistry(f *fetcher.Fetcher, sourcesPath string) (*Registry, error) {
	configs, err := LoadSourceConfigs(sourcesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load source configs: %w", err)
	}

	scrapers := make([]SourceScraper, 0, len(configs))
	for _, cfg := range configs {
		scrapers = append(scrapers, NewGenericScraper(cfg, f))
	}

	return &Registry{scrapers: scrapers}, nil
}

func (r *Registry) GetScrapers() []SourceScraper {
	return r.scrapers
}

func (r *Registry) GetScraperByName(name string) SourceScraper {
	for _, scraper := range r.scrapers {
		if scraper.Name() == name {
			return scraper
		}
	}
	return nil
}

func (r *Registry) GetScrapersByLanguage(language string) []SourceScraper {
	var filtered []SourceScraper
	for _, scraper := range r.scrapers {
		for _, lang := range scraper.Languages() {
			if string(lang) == language {
				filtered = append(filtered, scraper)
				break
			}
		}
	}
	return filtered
}
