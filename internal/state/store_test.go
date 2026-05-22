package state

import (
	"path/filepath"
	"testing"
)

func TestStoreSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	store := New(dir)

	m := NewManifest()
	m.Upsert(Entry{
		Name:        "kubectl",
		Version:     "1.30.2",
		Source:      "builtin",
		URL:         "https://example.test/kubectl",
		InstallPath: filepath.Join(dir, "kubectl"),
		SHA256:      "abc",
	})

	if err := store.Save(m); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	entry, ok := loaded.Entries["kubectl"]
	if !ok {
		t.Fatalf("expected kubectl entry")
	}
	if entry.Version != "1.30.2" {
		t.Fatalf("expected version 1.30.2, got %q", entry.Version)
	}
}
