package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"ipmanlk/cnapi/internal/config"
	"ipmanlk/cnapi/internal/fetcher"
	"ipmanlk/cnapi/internal/model"
	"ipmanlk/cnapi/internal/scraper"
)

var (
	sourceFilter = flag.String("source", "", "filter by source name (e.g., hiru)")
)

func main() {
	flag.Parse()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger := setupLogger(cfg.Logger)
	slog.SetDefault(logger)

	httpClient := fetcher.NewHTTPClient(cfg.Fetcher.HTTPTimeout)
	browserClient := fetcher.NewBrowserAPIClient(
		cfg.Fetcher.BrowserAPIURL,
		cfg.Fetcher.BrowserTimeout,
		cfg.Fetcher.BrowserWaitTime,
	)
	f := fetcher.NewFetcher(httpClient, browserClient)

	registry, err := scraper.NewRegistry(f, cfg.SourcesPath)
	if err != nil {
		slog.Error("failed to create scraper registry", "error", err)
		os.Exit(1)
	}

	var scrapers []scraper.Scraper

	if *sourceFilter != "" {
		s := registry.GetScraperByName(*sourceFilter)
		if s == nil {
			fmt.Printf("Source %q not found\n", *sourceFilter)
			os.Exit(1)
		}
		scrapers = []scraper.Scraper{s}
	} else {
		scrapers = registry.GetScrapers()
	}

	fmt.Printf("Testing %d sources...\n\n", len(scrapers))

	var wg sync.WaitGroup
	results := make(chan result, len(scrapers)*3)

	for _, s := range scrapers {
		for _, lang := range s.Languages() {
			wg.Add(1)
			go testScraper(ctx, s, lang, results, &wg)
		}
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var passed int
	var failed []result
	for r := range results {
		if r.success {
			passed++
			fmt.Printf("✓ %s [%s]: %d articles\n", r.source, r.language, r.count)
		} else {
			failed = append(failed, r)
		}
	}

	fmt.Printf("\n%d passed, %d failed\n", passed, len(failed))

	if len(failed) > 0 {
		fmt.Println("\nFailed sources:")
		for _, f := range failed {
			fmt.Printf("  ✗ %s [%s]: %s\n", f.source, f.language, f.error)
		}
		os.Exit(1)
	}
}

type result struct {
	source   string
	language model.Language
	success  bool
	count    int
	error    string
}

func testScraper(ctx context.Context, s scraper.Scraper, lang model.Language, results chan<- result, wg *sync.WaitGroup) {
	defer wg.Done()

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	articles, err := s.Scrape(ctx, lang)
	if err != nil {
		results <- result{
			source:   s.Name(),
			language: lang,
			success:  false,
			error:    err.Error(),
		}
		return
	}

	if len(articles) == 0 {
		results <- result{
			source:   s.Name(),
			language: lang,
			success:  false,
			error:    "no articles returned",
		}
		return
	}

	results <- result{
		source:   s.Name(),
		language: lang,
		success:  true,
		count:    len(articles),
	}
}

func setupLogger(cfg config.LoggerConfig) *slog.Logger {
	var level slog.Level
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: cfg.AddSource,
	}

	var handler slog.Handler
	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}
