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
	seenIDs := make(map[string]string) // id -> filename for duplicate detection

	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != ".toml" {
			continue
		}

		path := filepath.Join(sourcesPath, entry.Name())
		cfg, err := LoadConfig(path)
		if err != nil {
			slog.Error("failed to load source file, skipping", "file", entry.Name(), "error", err)
			continue
		}

		// Check for duplicate IDs
		if existingFile, exists := seenIDs[cfg.ID]; exists {
			slog.Error("duplicate source id detected, skipping",
				"file", entry.Name(),
				"id", cfg.ID,
				"existing_file", existingFile,
			)
			continue
		}
		seenIDs[cfg.ID] = entry.Name()

		configs = append(configs, *cfg)
	}

	if len(configs) == 0 {
		return nil, fmt.Errorf("no valid source configurations found in %s", sourcesPath)
	}

	slog.Info("loaded source configurations", "count", len(configs))
	return configs, nil
}
