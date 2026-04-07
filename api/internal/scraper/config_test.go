package scraper

import (
	"fmt"
	"path/filepath"
	"testing"
)

func TestLoadAllSourceConfigs(t *testing.T) {
	sourcesDir := "../../sources"

	// Find all TOML files
	tomlFiles, err := filepath.Glob(filepath.Join(sourcesDir, "*.toml"))
	if err != nil {
		t.Fatalf("Failed to glob sources directory: %v", err)
	}

	if len(tomlFiles) == 0 {
		t.Fatal("No TOML files found in sources directory")
	}

	t.Logf("Found %d source config files", len(tomlFiles))

	// Load each config file
	var errors []error
	for _, file := range tomlFiles {
		name := filepath.Base(file)
		t.Run(name, func(t *testing.T) {
			cfg, err := LoadConfig(file)
			if err != nil {
				errors = append(errors, fmt.Errorf("%s: %w", name, err))
				t.Errorf("Failed to load config: %v", err)
				return
			}

			// Basic validation
			if cfg.Name == "" {
				errors = append(errors, fmt.Errorf("%s: missing name field", name))
				t.Error("Config missing name field")
				return
			}

			if len(cfg.Languages) == 0 {
				errors = append(errors, fmt.Errorf("%s: no languages defined", name))
				t.Error("Config has no languages defined")
				return
			}

			// Validate each language
			for _, lang := range cfg.Languages {
				if lang.Language == "" {
					errors = append(errors, fmt.Errorf("%s: language missing language code", name))
					t.Error("Language config missing language code")
					continue
				}
				if lang.Discovery.Type == "" {
					errors = append(errors, fmt.Errorf("%s [%s]: discovery type not set", name, lang.Language))
					t.Errorf("Language %s: discovery type not set", lang.Language)
				}
			}

			t.Logf("✓ %s loaded successfully (%d languages)", name, len(cfg.Languages))
		})
	}

	if len(errors) > 0 {
		t.Fatalf("\n%d config(s) failed to load:", len(errors))
	}
}

func TestConfigLoading(t *testing.T) {
	// Test loading a specific config
	sourcesDir := "../../sources"
	tomlFiles, _ := filepath.Glob(filepath.Join(sourcesDir, "*.toml"))

	if len(tomlFiles) == 0 {
		t.Skip("No TOML files to test")
	}

	// Test the first file as a sample
	testFile := tomlFiles[0]
	cfg, err := LoadConfig(testFile)
	if err != nil {
		t.Fatalf("Failed to load sample config %s: %v", testFile, err)
	}

	t.Logf("Sample config '%s' loaded:", cfg.Name)
	t.Logf("  - Languages: %d", len(cfg.Languages))
	for _, lang := range cfg.Languages {
		t.Logf("    - %s (max_items: %d)", lang.Language, lang.MaxItems)
	}
}
