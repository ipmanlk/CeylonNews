package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Database    DatabaseConfig
	Fetcher     FetcherConfig
	Logger      LoggerConfig
	Scheduler   SchedulerConfig
	HTTP        HTTPConfig
	SourcesPath string
}

type SchedulerConfig struct {
	ScrapeInterval time.Duration
	Enabled        bool
	HTTPWorkers    int
	BrowserWorkers int
}

type DatabaseConfig struct {
	Driver          string
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type FetcherConfig struct {
	HTTPTimeout     time.Duration
	BrowserAPIURL   string
	BrowserTimeout  time.Duration
	BrowserWaitTime int
}

type LoggerConfig struct {
	Level     string
	Format    string
	AddSource bool
}

type HTTPConfig struct {
	Host            string
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

func Load() (*Config, error) {
	cfg := &Config{
		Database: DatabaseConfig{
			Driver:          getEnv("DB_DRIVER", "sqlite3"),
			DSN:             getEnv("DB_DSN", "./data/db.sqlite"),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		Fetcher: FetcherConfig{
			HTTPTimeout:     getEnvDuration("FETCHER_HTTP_TIMEOUT", 15*time.Second),
			BrowserAPIURL:   getEnv("FETCHER_BROWSER_API_URL", "http://localhost:8000"),
			BrowserTimeout:  getEnvDuration("FETCHER_BROWSER_TIMEOUT", 60*time.Second),
			BrowserWaitTime: getEnvInt("FETCHER_BROWSER_WAIT_TIME", 15),
		},
		Logger: LoggerConfig{
			Level:     getEnv("LOG_LEVEL", "info"),
			Format:    getEnv("LOG_FORMAT", "json"),
			AddSource: getEnvBool("LOG_ADD_SOURCE", true),
		},
		Scheduler: SchedulerConfig{
			ScrapeInterval: getEnvDuration("SCHEDULER_SCRAPE_INTERVAL", 1*time.Hour),
			Enabled:        getEnvBool("SCHEDULER_ENABLED", true),
			HTTPWorkers:    getEnvInt("SCHEDULER_HTTP_WORKERS", 4),
			BrowserWorkers: getEnvInt("SCHEDULER_BROWSER_WORKERS", 1),
		},
		HTTP: HTTPConfig{
			Host:            getEnv("HTTP_HOST", "0.0.0.0"),
			Port:            getEnvInt("HTTP_PORT", 8080),
			ReadTimeout:     getEnvDuration("HTTP_READ_TIMEOUT", 30*time.Second),
			WriteTimeout:    getEnvDuration("HTTP_WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:     getEnvDuration("HTTP_IDLE_TIMEOUT", 120*time.Second),
			ShutdownTimeout: getEnvDuration("HTTP_SHUTDOWN_TIMEOUT", 15*time.Second),
		},
		SourcesPath: getEnv("SOURCES_PATH", "./sources"),
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.Database.DSN == "" {
		return fmt.Errorf("database DSN is required")
	}

	if c.Fetcher.BrowserAPIURL == "" {
		return fmt.Errorf("browser API URL is required")
	}

	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLogLevels[c.Logger.Level] {
		return fmt.Errorf("invalid log level: %s", c.Logger.Level)
	}

	if c.Logger.Format != "json" && c.Logger.Format != "text" {
		return fmt.Errorf("invalid log format: %s (must be 'json' or 'text')", c.Logger.Format)
	}

	if c.Scheduler.Enabled && c.Scheduler.ScrapeInterval < 1*time.Minute {
		return fmt.Errorf("scrape interval must be at least 1 minute, got: %v", c.Scheduler.ScrapeInterval)
	}

	if c.HTTP.Port < 1 || c.HTTP.Port > 65535 {
		return fmt.Errorf("invalid HTTP port: %d (must be between 1 and 65535)", c.HTTP.Port)
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

func getEnvBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	if valueStr == "0" {
		return 0
	}

	value, err := time.ParseDuration(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}
