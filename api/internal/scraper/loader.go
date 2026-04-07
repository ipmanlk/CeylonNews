package scraper

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

func LoadConfigs(sourcesPath string) ([]Config, error) {
	entries, err := os.ReadDir(sourcesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read sources directory: %w", err)
	}

	var configs []Config
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != ".toml" {
			continue
		}

		path := filepath.Join(sourcesPath, entry.Name())
		cfg, err := LoadConfig(path)
		if err != nil {
			slog.Warn("failed to load source file", "file", entry.Name(), "error", err)
			continue
		}

		configs = append(configs, *cfg)
	}

	return configs, nil
}
