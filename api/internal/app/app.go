package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	"ipmanlk/cnapi/internal/api"
	"ipmanlk/cnapi/internal/config"
	"ipmanlk/cnapi/internal/database"
	"ipmanlk/cnapi/internal/fetcher"
	"ipmanlk/cnapi/internal/scheduler"
	"ipmanlk/cnapi/internal/scraper"
	"ipmanlk/cnapi/internal/service"
)

type App struct {
	Config     *config.Config
	DB         *sql.DB
	Store      *database.Store
	Scheduler  *scheduler.Scheduler
	HTTPServer *api.Server
	Logger     *slog.Logger
}

func New(ctx context.Context) (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	logger := setupLogger(cfg.Logger)
	slog.SetDefault(logger)

	slog.Info("starting application initialization")

	db, err := initDatabase(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	if err := database.InitializeDatabase(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run database migrations: %w", err)
	}

	store := database.NewStore(db)

	httpClient := fetcher.NewHTTPClient(cfg.Fetcher.HTTPTimeout)
	browserClient := fetcher.NewBrowserAPIClient(
		cfg.Fetcher.BrowserAPIURL,
		cfg.Fetcher.BrowserTimeout,
		cfg.Fetcher.BrowserWaitTime,
	)
	fetch := fetcher.NewFetcher(httpClient, browserClient)

	scraperRegistry, err := scraper.NewRegistry(fetch, cfg.SourcesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize scraper registry: %w", err)
	}

	scrapeService := service.NewScrapeService(scraperRegistry)
	articleService := service.NewArticleService(store.Articles)
	searchService := service.NewSearchService(store.Search)

	sched := scheduler.New(
		scrapeService,
		articleService,
		cfg.Scheduler.ScrapeInterval,
		cfg.Scheduler.HTTPWorkers,
		cfg.Scheduler.BrowserWorkers,
	)

	httpConfig := api.Config{
		Host:            cfg.HTTP.Host,
		Port:            cfg.HTTP.Port,
		ReadTimeout:     cfg.HTTP.ReadTimeout,
		WriteTimeout:    cfg.HTTP.WriteTimeout,
		IdleTimeout:     cfg.HTTP.IdleTimeout,
		ShutdownTimeout: cfg.HTTP.ShutdownTimeout,
	}
	httpServer := api.NewServer(articleService, searchService, httpConfig)

	app := &App{
		Config:     cfg,
		DB:         db,
		Store:      store,
		Scheduler:  sched,
		HTTPServer: httpServer,
		Logger:     logger,
	}

	slog.Info("application initialized successfully")

	return app, nil
}

func (a *App) Close(ctx context.Context) error {
	slog.Info("shutting down application")

	if a.HTTPServer != nil {
		if err := a.HTTPServer.Shutdown(ctx); err != nil {
			slog.Error("error stopping HTTP server", "error", err)
		}
	}

	if a.Scheduler != nil && a.Scheduler.IsRunning() {
		if err := a.Scheduler.Stop(); err != nil {
			slog.Error("error stopping scheduler", "error", err)
		}
	}

	if a.DB != nil {
		if err := a.DB.Close(); err != nil {
			return fmt.Errorf("failed to close database: %w", err)
		}
	}

	slog.Info("application shutdown complete")
	return nil
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

func initDatabase(cfg config.DatabaseConfig) (*sql.DB, error) {
	db, err := sql.Open(cfg.Driver, cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	slog.Info("database connection established",
		"driver", cfg.Driver,
		"max_open_conns", cfg.MaxOpenConns,
		"max_idle_conns", cfg.MaxIdleConns,
	)

	return db, nil
}
