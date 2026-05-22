package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	toml "github.com/pelletier/go-toml/v2"
)

func ParseFile(path string) (Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config %s: %w", path, err)
	}
	cfg, err := Parse(content)
	if err != nil {
		return Config{}, fmt.Errorf("parse config %s: %w", path, err)
	}
	return cfg, nil
}

func Parse(data []byte) (Config, error) {
	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}

	applyDefaults(&cfg)
	if err := validate(cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func applyDefaults(cfg *Config) {
	if cfg.Version == 0 {
		cfg.Version = 1
	}

	for i := range cfg.Tools {
		cfg.Tools[i].Name = strings.TrimSpace(cfg.Tools[i].Name)
		cfg.Tools[i].Source = strings.ToLower(strings.TrimSpace(cfg.Tools[i].Source))
		cfg.Tools[i].Version = strings.TrimSpace(cfg.Tools[i].Version)
		cfg.Tools[i].URL = strings.TrimSpace(cfg.Tools[i].URL)
		cfg.Tools[i].Type = strings.ToLower(strings.TrimSpace(cfg.Tools[i].Type))
		cfg.Tools[i].ArchiveBinary = strings.TrimSpace(cfg.Tools[i].ArchiveBinary)
		cfg.Tools[i].BinaryName = strings.TrimSpace(cfg.Tools[i].BinaryName)
		cfg.Tools[i].SHA256 = strings.TrimSpace(cfg.Tools[i].SHA256)

		if cfg.Tools[i].Source == "" {
			cfg.Tools[i].Source = SourceBuiltin
		}
		if cfg.Tools[i].BinaryName == "" {
			cfg.Tools[i].BinaryName = cfg.Tools[i].Name
		}
		if cfg.Tools[i].Source == SourceCustom && cfg.Tools[i].Type == "" {
			switch {
			case strings.HasSuffix(cfg.Tools[i].URL, ".tar.gz"),
				strings.HasSuffix(cfg.Tools[i].URL, ".tgz"),
				strings.HasSuffix(cfg.Tools[i].URL, ".zip"):
				cfg.Tools[i].Type = ToolTypeArchive
			default:
				cfg.Tools[i].Type = ToolTypeBinary
			}
		}
	}
}

func validate(cfg Config) error {
	if cfg.Version != 1 {
		return fmt.Errorf("unsupported config version %d", cfg.Version)
	}

	for i, t := range cfg.Tools {
		idx := i + 1
		if t.Name == "" {
			return fmt.Errorf("tools[%d]: name is required", idx)
		}

		if t.Source != SourceBuiltin && t.Source != SourceCustom {
			return fmt.Errorf("tools[%d]: source must be %q or %q", idx, SourceBuiltin, SourceCustom)
		}

		if t.BinaryName != filepath.Base(t.BinaryName) {
			return fmt.Errorf("tools[%d]: binary_name must be a file name, not a path", idx)
		}

		if t.Source == SourceCustom {
			if t.URL == "" {
				return fmt.Errorf("tools[%d]: url is required for custom tools", idx)
			}
			if t.Type != ToolTypeBinary && t.Type != ToolTypeArchive {
				return fmt.Errorf("tools[%d]: type must be %q or %q", idx, ToolTypeBinary, ToolTypeArchive)
			}
			if t.Type == ToolTypeArchive && t.ArchiveBinary == "" {
				return fmt.Errorf("tools[%d]: archive_binary is required for archive tools", idx)
			}
		}
	}
	return nil
}
