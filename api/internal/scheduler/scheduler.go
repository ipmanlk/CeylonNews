package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"ipmanlk/cnapi/internal/model"
	"ipmanlk/cnapi/internal/service"
)

type ScrapeService interface {
	ScrapeAllConcurrent(ctx context.Context, httpWorkers int, browserWorkers int, batchSize int) <-chan service.ScrapeResult
}

type ArticleService interface {
	BulkUpsert(ctx context.Context, articles []model.ScrapedArticle) ([]int64, error)
}

type Scheduler struct {
	scrapeService  ScrapeService
	articleService ArticleService
	interval       time.Duration
	httpWorkers    int
	browserWorkers int
	ticker         *time.Ticker
	stopCh         chan struct{}
	doneCh         chan struct{}
	running        bool
	mu             sync.RWMutex
	nextRun        time.Time
}

func New(scrapeService ScrapeService, articleService ArticleService, interval time.Duration, httpWorkers int, browserWorkers int) *Scheduler {
	return &Scheduler{
		scrapeService:  scrapeService,
		articleService: articleService,
		interval:       interval,
		httpWorkers:    httpWorkers,
		browserWorkers: browserWorkers,
		running:        false,
	}
}

func (s *Scheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("scheduler is already running")
	}

	slog.Info("starting scheduler", "interval", s.interval)

	s.ticker = time.NewTicker(s.interval)
	s.stopCh = make(chan struct{})
	s.doneCh = make(chan struct{})
	s.running = true
	s.nextRun = time.Now().Add(s.interval)
	s.mu.Unlock()

	go s.run(ctx)

	slog.Info("scheduler started successfully", "next_run", s.getNextRunTime())

	go s.scrapeAndStoreAll(ctx)

	return nil
}

func (s *Scheduler) run(ctx context.Context) {
	defer close(s.doneCh)

	for {
		select {
		case <-s.ticker.C:
			s.mu.Lock()
			s.nextRun = time.Now().Add(s.interval)
			s.mu.Unlock()
			s.scrapeAndStoreAll(ctx)
		case <-s.stopCh:
			s.ticker.Stop()
			return
		case <-ctx.Done():
			s.ticker.Stop()
			return
		}
	}
}

func (s *Scheduler) Stop() error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return nil
	}
	s.mu.Unlock()

	slog.Info("stopping scheduler")

	close(s.stopCh)

	<-s.doneCh

	s.mu.Lock()
	s.running = false
	s.mu.Unlock()

	slog.Info("scheduler stopped")

	return nil
}

func (s *Scheduler) scrapeAndStoreAll(ctx context.Context) {
	startTime := time.Now()
	slog.Info("starting scheduled scrape job")

	// Create a context with timeout for this job
	jobCtx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	// Use concurrent scraping with streaming
	// 4 workers = 4 concurrent scrapers at a time
	// 100 batch size = insert every 100 articles
	resultChan := s.scrapeService.ScrapeAllConcurrent(jobCtx, s.httpWorkers, s.browserWorkers, 100)

	totalScraped := 0
	totalStored := 0
	successCount := 0
	failureCount := 0
	storeStartTime := time.Now()

	// Process results as they come in
	for result := range resultChan {
		if !result.Success {
			failureCount++
			continue
		}

		successCount++

		if len(result.Articles) == 0 {
			continue
		}

		totalScraped += len(result.Articles)

		// Store this batch immediately
		ids, err := s.articleService.BulkUpsert(jobCtx, result.Articles)
		if err != nil {
			slog.Error("failed to store articles batch",
				"error", err,
				"source", result.Source,
				"language", result.Language,
				"article_count", len(result.Articles),
			)
			continue
		}

		totalStored += len(ids)

		slog.Debug("stored articles batch",
			"source", result.Source,
			"language", result.Language,
			"scraped", len(result.Articles),
			"stored", len(ids),
		)
	}

	storeDuration := time.Since(storeStartTime)
	totalDuration := time.Since(startTime)

	if totalScraped == 0 {
		slog.Warn("no articles scraped",
			"duration", totalDuration,
			"successful_sources", successCount,
			"failed_sources", failureCount,
		)
		return
	}

	slog.Info("scrape and store job completed successfully",
		"total_scraped", totalScraped,
		"stored_count", totalStored,
		"successful_sources", successCount,
		"failed_sources", failureCount,
		"total_duration", totalDuration.Round(time.Second),
		"store_duration", storeDuration.Round(time.Second),
		"next_run", s.getNextRunTime(),
	)
}

func (s *Scheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

func (s *Scheduler) getNextRunTime() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.running {
		return "scheduler not running"
	}
	return s.nextRun.Format(time.RFC3339)
}

func (s *Scheduler) RunNow(ctx context.Context) {
	slog.Info("manual scrape job triggered")
	go s.scrapeAndStoreAll(ctx)
}
