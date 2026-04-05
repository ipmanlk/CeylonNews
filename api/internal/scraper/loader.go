package scraper

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

func LoadSourceConfigs(sourcesPath string) ([]SourceConfig, error) {
	entries, err := os.ReadDir(sourcesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read sources directory: %w", err)
	}

	var configs []SourceConfig
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != ".toml" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(sourcesPath, entry.Name()))
		if err != nil {
			slog.Warn("failed to read source file", "file", entry.Name(), "error", err)
			continue
		}

		var cfg SourceConfig
		if err := toml.Unmarshal(data, &cfg); err != nil {
			slog.Warn("failed to parse source file", "file", entry.Name(), "error", err)
			continue
		}

		configs = append(configs, cfg)
	}

	return configs, nil
}
