package state

import (
	"sort"
	"time"
)

type Entry struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Source      string `json:"source"`
	URL         string `json:"url"`
	InstallPath string `json:"install_path"`
	SHA256      string `json:"sha256"`
	UpdatedAt   string `json:"updated_at"`
}

type Manifest struct {
	Entries map[string]Entry `json:"entries"`
}

func NewManifest() Manifest {
	return Manifest{Entries: map[string]Entry{}}
}

func (m *Manifest) Upsert(entry Entry) {
	if m.Entries == nil {
		m.Entries = map[string]Entry{}
	}
	if entry.UpdatedAt == "" {
		entry.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	}
	m.Entries[entry.Name] = entry
}

func (m *Manifest) Remove(name string) (Entry, bool) {
	if m.Entries == nil {
		return Entry{}, false
	}
	current, ok := m.Entries[name]
	if ok {
		delete(m.Entries, name)
	}
	return current, ok
}

func (m Manifest) Sorted() []Entry {
	if len(m.Entries) == 0 {
		return nil
	}
	out := make([]Entry, 0, len(m.Entries))
	for _, entry := range m.Entries {
		out = append(out, entry)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Name < out[j].Name
	})
	return out
}
