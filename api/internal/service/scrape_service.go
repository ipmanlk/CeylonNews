package service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"ipmanlk/cnapi/internal/model"
	"ipmanlk/cnapi/internal/scraper"
)

type ScrapeResult struct {
	Articles     []model.ScrapedArticle
	Source       string
	Language     model.Language
	ArticleCount int
	Success      bool
	Error        error
}

type scrapeService struct {
	registry *scraper.Registry
}

func NewScrapeService(registry *scraper.Registry) *scrapeService {
	return &scrapeService{registry: registry}
}

type scrapeTask struct {
	scraper  scraper.Scraper
	language model.Language
}

// ScrapeAllConcurrent fans out scraping across all sources using separate HTTP
// and browser worker pools. Results are streamed on the returned channel, which
// is closed when all workers finish.
func (s *scrapeService) ScrapeAllConcurrent(ctx context.Context, httpWorkers, browserWorkers, batchSize int) <-chan ScrapeResult {
	resultChan := make(chan ScrapeResult, httpWorkers+browserWorkers)

	var httpTasks []scrapeTask
	var browserTasks []scrapeTask

	scrapers := s.registry.GetScrapers()
	for _, sc := range scrapers {
		for _, lang := range sc.Languages() {
			task := scrapeTask{scraper: sc, language: lang}
			if sc.UsesBrowser(lang) {
				browserTasks = append(browserTasks, task)
			} else {
				httpTasks = append(httpTasks, task)
			}
		}
	}

	go func() {
		defer close(resultChan)

		var wg sync.WaitGroup

		httpCh := make(chan scrapeTask, len(httpTasks))
		for i := 0; i < httpWorkers; i++ {
			wg.Add(1)
			go s.worker(ctx, httpCh, resultChan, batchSize, &wg)
		}

		browserCh := make(chan scrapeTask, len(browserTasks))
		for i := 0; i < browserWorkers; i++ {
			wg.Add(1)
			go s.worker(ctx, browserCh, resultChan, batchSize, &wg)
		}

		for _, task := range httpTasks {
			select {
			case httpCh <- task:
			case <-ctx.Done():
				close(httpCh)
				close(browserCh)
				wg.Wait()
				return
			}
		}
		close(httpCh)

		for _, task := range browserTasks {
			select {
			case browserCh <- task:
			case <-ctx.Done():
				close(browserCh)
				wg.Wait()
				return
			}
		}
		close(browserCh)

		wg.Wait()
	}()

	return resultChan
}

func (s *scrapeService) worker(ctx context.Context, tasks <-chan scrapeTask, results chan<- ScrapeResult, batchSize int, wg *sync.WaitGroup) {
	defer wg.Done()

	for task := range tasks {
		select {
		case <-ctx.Done():
			return
		default:
			articles, err := task.scraper.Scrape(ctx, task.language)

			result := ScrapeResult{
				Source:   task.scraper.Name(),
				Language: task.language,
				Success:  err == nil,
				Error:    err,
			}

			if err != nil {
				slog.Warn("failed to scrape from source",
					"source", task.scraper.Name(),
					"language", task.language,
					"error", err,
				)
				result.ArticleCount = 0
				result.Articles = nil
				results <- result
				continue
			}

			slog.Info("successfully scraped articles",
				"source", task.scraper.Name(),
				"language", task.language,
				"count", len(articles),
			)

			if batchSize <= 0 || len(articles) <= batchSize {
				result.Articles = articles
				result.ArticleCount = len(articles)
				results <- result
			} else {
				for i := 0; i < len(articles); i += batchSize {
					end := i + batchSize
					if end > len(articles) {
						end = len(articles)
					}

					batchResult := ScrapeResult{
						Source:       task.scraper.Name(),
						Language:     task.language,
						Articles:     articles[i:end],
						ArticleCount: end - i,
						Success:      true,
					}

					select {
					case results <- batchResult:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}
}

func (s *scrapeService) ScrapeBySource(ctx context.Context, sourceID string) ([]model.ScrapedArticle, error) {
	sc := s.registry.GetScraperByID(sourceID)
	if sc == nil {
		return nil, fmt.Errorf("scraper not found: %s", sourceID)
	}

	var allArticles []model.ScrapedArticle
	for _, lang := range sc.Languages() {
		articles, err := sc.Scrape(ctx, lang)
		if err != nil {
			slog.Warn("failed to scrape from source",
				"source", sourceID,
				"language", lang,
				"error", err,
			)
			continue
		}
		allArticles = append(allArticles, articles...)
	}

	return allArticles, nil
}

func (s *scrapeService) ScrapeByLanguage(ctx context.Context, language model.Language) ([]model.ScrapedArticle, error) {
	var allArticles []model.ScrapedArticle

	for _, sc := range s.registry.GetScrapersByLanguage(string(language)) {
		articles, err := sc.Scrape(ctx, language)
		if err != nil {
			slog.Warn("failed to scrape from source",
				"source", sc.Name(),
				"language", language,
				"error", err,
			)
			continue
		}
		allArticles = append(allArticles, articles...)
	}

	return allArticles, nil
}

func (s *scrapeService) GetAvailableSources() []string {
	scrapers := s.registry.GetScrapers()
	sources := make([]string, len(scrapers))
	for i, sc := range scrapers {
		sources[i] = sc.ID()
	}
	return sources
}

func (s *scrapeService) GetAvailableLanguages() []model.Language {
	seen := make(map[model.Language]bool)
	for _, sc := range s.registry.GetScrapers() {
		for _, lang := range sc.Languages() {
			seen[lang] = true
		}
	}

	languages := make([]model.Language, 0, len(seen))
	for lang := range seen {
		languages = append(languages, lang)
	}
	return languages
}
