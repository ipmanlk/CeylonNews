package service

import (
	"context"

	"ipmanlk/cnapi/internal/database"
	"ipmanlk/cnapi/internal/database/store"
	"ipmanlk/cnapi/internal/model"
	"ipmanlk/cnapi/internal/scraper"
)

type searchService struct {
	store           database.SearchStore
	scraperRegistry *scraper.Registry
}

func NewSearchService(store database.SearchStore, registry *scraper.Registry) *searchService {
	return &searchService{store: store, scraperRegistry: registry}
}

func (s *searchService) Search(ctx context.Context, filter model.SearchFilter) (*model.Paginated[*model.SearchResult], error) {
	return s.store.Search(filter)
}

func (s *searchService) GetAvailableSources() ([]string, error) {
	return s.store.GetAvailableSources()
}

func (s *searchService) GetAvailableLanguages() ([]string, error) {
	return s.store.GetAvailableLanguages()
}

func (s *searchService) GetSourcesByLanguage(language string) ([]store.SourceInfo, error) {
	sources, err := s.store.GetSourcesByLanguage(language)
	if err != nil {
		return nil, err
	}
	for i := range sources {
		if name, ok := s.scraperRegistry.GetSourceNameByID(sources[i].ID); ok {
			sources[i].Name = name
		}
	}
	return sources, nil
}

func (s *searchService) GetRecentArticles(languages []string, sourceIDs []string, limit int) ([]*model.Article, error) {
	return s.store.GetRecentArticles(languages, sourceIDs, limit)
}
