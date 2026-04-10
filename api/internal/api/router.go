package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"ipmanlk/cnapi/internal/api/handlers"
	"ipmanlk/cnapi/internal/api/middleware"
)

type Server struct {
	httpServer *http.Server
	config     Config
}

type Config struct {
	Host            string
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

func NewServer(articleService handlers.ArticleService, searchService handlers.SearchService, sourceResolver handlers.SourceResolver, config Config) *Server {
	mux := http.NewServeMux()

	articleHandler := handlers.NewArticleHandler(articleService, sourceResolver)
	searchHandler := handlers.NewSearchHandler(searchService, sourceResolver)

	mux.HandleFunc("GET /health", handlers.Health)

	mux.HandleFunc("GET /api/v1/articles", articleHandler.List)
	mux.HandleFunc("GET /api/v1/articles/{id}", articleHandler.GetByID)

	mux.HandleFunc("GET /api/v1/search", searchHandler.Search)
	mux.HandleFunc("GET /api/v1/search/sources", searchHandler.GetAvailableSources)
	mux.HandleFunc("GET /api/v1/search/languages", searchHandler.GetAvailableLanguages)
	mux.HandleFunc("GET /api/v1/search/sources/by-language", searchHandler.GetSourcesByLanguage)
	mux.HandleFunc("GET /api/v1/search/recent", searchHandler.GetRecentArticles)

	handler := middleware.Logging(
		middleware.Recovery(
			middleware.CORS(mux),
		),
	)

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", config.Host, config.Port),
		Handler: handler,
	}

	if config.ReadTimeout > 0 {
		srv.ReadTimeout = config.ReadTimeout
	}
	if config.WriteTimeout > 0 {
		srv.WriteTimeout = config.WriteTimeout
	}
	if config.IdleTimeout > 0 {
		srv.IdleTimeout = config.IdleTimeout
	}

	return &Server{
		httpServer: srv,
		config:     config,
	}
}

func (s *Server) Start() error {
	slog.Info("starting HTTP server", "address", s.httpServer.Addr)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	slog.Info("shutting down HTTP server")

	shutdownCtx, cancel := context.WithTimeout(ctx, s.config.ShutdownTimeout)
	defer cancel()

	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	slog.Info("HTTP server shutdown complete")
	return nil
}
