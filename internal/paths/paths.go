package paths

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/robert/toolbox/internal/config"
)

type Dirs struct {
	BinDir   string
	StateDir string
	CacheDir string
}

func Resolve(defaults config.Defaults, binOverride string) (Dirs, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Dirs{}, fmt.Errorf("resolve user home: %w", err)
	}

	binDir := strings.TrimSpace(defaults.BinDir)
	if binDir == "" {
		binDir = filepath.Join(home, ".local", "bin")
	}
	if strings.TrimSpace(binOverride) != "" {
		binDir = binOverride
	}

	stateDir := strings.TrimSpace(defaults.StateDir)
	if stateDir == "" {
		if xdgData := strings.TrimSpace(os.Getenv("XDG_DATA_HOME")); xdgData != "" {
			stateDir = filepath.Join(xdgData, "toolbox")
		} else {
			stateDir = filepath.Join(home, ".local", "share", "toolbox")
		}
	}

	cacheDir := strings.TrimSpace(defaults.CacheDir)
	if cacheDir == "" {
		if xdgCache := strings.TrimSpace(os.Getenv("XDG_CACHE_HOME")); xdgCache != "" {
			cacheDir = filepath.Join(xdgCache, "toolbox")
		} else {
			cacheDir = filepath.Join(home, ".cache", "toolbox")
		}
	}

	binDir, err = expandHome(binDir, home)
	if err != nil {
		return Dirs{}, fmt.Errorf("resolve bin_dir: %w", err)
	}
	stateDir, err = expandHome(stateDir, home)
	if err != nil {
		return Dirs{}, fmt.Errorf("resolve state_dir: %w", err)
	}
	cacheDir, err = expandHome(cacheDir, home)
	if err != nil {
		return Dirs{}, fmt.Errorf("resolve cache_dir: %w", err)
	}

	return Dirs{
		BinDir:   filepath.Clean(binDir),
		StateDir: filepath.Clean(stateDir),
		CacheDir: filepath.Clean(cacheDir),
	}, nil
}

func Ensure(d Dirs) error {
	for _, dir := range []string{d.BinDir, d.StateDir, d.CacheDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create %s: %w", dir, err)
		}
	}
	return nil
}

func expandHome(path, home string) (string, error) {
	trimmed := strings.TrimSpace(path)
	switch {
	case trimmed == "~":
		return home, nil
	case strings.HasPrefix(trimmed, "~/"):
		return filepath.Join(home, strings.TrimPrefix(trimmed, "~/")), nil
	}
	return trimmed, nil
}
