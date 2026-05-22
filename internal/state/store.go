package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type Store struct {
	FilePath string
}

func New(stateDir string) Store {
	return Store{
		FilePath: filepath.Join(stateDir, "state.json"),
	}
}

func (s Store) Load() (Manifest, error) {
	content, err := os.ReadFile(s.FilePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return NewManifest(), nil
		}
		return Manifest{}, fmt.Errorf("read state %s: %w", s.FilePath, err)
	}
	var manifest Manifest
	if err := json.Unmarshal(content, &manifest); err != nil {
		return Manifest{}, fmt.Errorf("parse state %s: %w", s.FilePath, err)
	}
	if manifest.Entries == nil {
		manifest.Entries = map[string]Entry{}
	}
	return manifest, nil
}

func (s Store) Save(manifest Manifest) error {
	if manifest.Entries == nil {
		manifest.Entries = map[string]Entry{}
	}
	if err := os.MkdirAll(filepath.Dir(s.FilePath), 0o755); err != nil {
		return fmt.Errorf("create state directory: %w", err)
	}

	content, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("encode state: %w", err)
	}

	tmpPath := s.FilePath + ".tmp"
	if err := os.WriteFile(tmpPath, content, 0o644); err != nil {
		return fmt.Errorf("write state temp file: %w", err)
	}
	if err := os.Rename(tmpPath, s.FilePath); err != nil {
		return fmt.Errorf("replace state file: %w", err)
	}
	return nil
}
