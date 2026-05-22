package doctor

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/robert/toolbox/internal/state"
)

type Report struct {
	Errors   []string
	Warnings []string
	Info     []string
}

func (r *Report) AddError(msg string) {
	r.Errors = append(r.Errors, msg)
}

func (r *Report) AddWarning(msg string) {
	r.Warnings = append(r.Warnings, msg)
}

func (r *Report) AddInfo(msg string) {
	r.Info = append(r.Info, msg)
}

func Run(configPath, binDir string, store state.Store) (Report, error) {
	var report Report

	if _, err := os.Stat(configPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			report.AddWarning(fmt.Sprintf("config file not found: %s", configPath))
		} else {
			report.AddError(fmt.Sprintf("cannot read config: %v", err))
		}
	}

	if err := ensureWritableDirectory(binDir); err != nil {
		report.AddError(fmt.Sprintf("install directory issue: %v", err))
	} else {
		report.AddInfo(fmt.Sprintf("install directory writable: %s", binDir))
	}

	if !pathContains(binDir, os.Getenv("PATH")) {
		report.AddWarning(fmt.Sprintf("PATH does not include %s", binDir))
	} else {
		report.AddInfo(fmt.Sprintf("PATH includes %s", binDir))
	}

	manifest, err := store.Load()
	if err != nil {
		return report, fmt.Errorf("load state: %w", err)
	}

	for _, entry := range manifest.Sorted() {
		if _, err := os.Stat(entry.InstallPath); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				report.AddWarning(fmt.Sprintf("managed tool missing on disk: %s (%s)", entry.Name, entry.InstallPath))
				continue
			}
			report.AddError(fmt.Sprintf("cannot inspect %s: %v", entry.Name, err))
		}
	}

	if len(manifest.Entries) == 0 {
		report.AddInfo("no managed tools currently recorded")
	}
	return report, nil
}

func ensureWritableDirectory(path string) error {
	if err := os.MkdirAll(path, 0o755); err != nil {
		return err
	}
	testPath := filepath.Join(path, ".toolbox-write-test")
	if err := os.WriteFile(testPath, []byte("ok"), 0o600); err != nil {
		return err
	}
	return os.Remove(testPath)
}

func pathContains(needle, pathValue string) bool {
	parts := strings.Split(pathValue, string(os.PathListSeparator))
	for _, part := range parts {
		if filepath.Clean(part) == filepath.Clean(needle) {
			return true
		}
	}
	return false
}
