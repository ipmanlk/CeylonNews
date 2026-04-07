package scraper

import (
	"context"
	"fmt"

	"ipmanlk/cnapi/internal/fetcher"
	"ipmanlk/cnapi/internal/model"
)

type Scraper interface {
	// Name returns the human-readable source name (e.g. "BBC", "Daily Mirror").
	Name() string

	// Languages returns the language codes this scraper can produce articles for.
	Languages() []model.Language

	// Scrape fetches and extracts articles for the given language.
	Scrape(ctx context.Context, language model.Language) ([]model.ScrapedArticle, error)

	// UsesBrowser reports whether scraping this language requires the headless
	// browser (used to fan tasks into the right worker pool).
	UsesBrowser(language model.Language) bool
}

type Registry struct {
	scrapers []Scraper
}

func NewRegistry(f *fetcher.Fetcher, sourcesPath string) (*Registry, error) {
	configs, err := LoadConfigs(sourcesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load source configs: %w", err)
	}

	scrapers := make([]Scraper, 0, len(configs))
	for _, cfg := range configs {
		scrapers = append(scrapers, NewSource(cfg, f))
	}

	return &Registry{scrapers: scrapers}, nil
}

func (r *Registry) GetScrapers() []Scraper {
	return r.scrapers
}

func (r *Registry) GetScraperByName(name string) Scraper {
	for _, s := range r.scrapers {
		if s.Name() == name {
			return s
		}
	}
	return nil
}

func (r *Registry) GetScrapersByLanguage(language string) []Scraper {
	var filtered []Scraper
	for _, s := range r.scrapers {
		for _, lang := range s.Languages() {
			if string(lang) == language {
				filtered = append(filtered, s)
				break
			}
		}
	}
	return filtered
}
