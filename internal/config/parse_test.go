package config

import "testing"

func TestParseDefaultsCustomArchiveType(t *testing.T) {
	input := []byte(`
version = 1

[[tools]]
name = "demo"
source = "custom"
version = "1.2.3"
url = "https://example.com/demo-{{os}}-{{arch}}.tar.gz"
archive_binaries = "demo"
`)

	cfg, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}
	if len(cfg.Tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(cfg.Tools))
	}
	if cfg.Tools[0].Type != ToolTypeArchive {
		t.Fatalf("expected inferred archive type, got %q", cfg.Tools[0].Type)
	}
	if cfg.Tools[0].BinaryName != "demo" {
		t.Fatalf("expected binary name fallback to tool name, got %q", cfg.Tools[0].BinaryName)
	}
	if !cfg.Defaults.AutoPathEnabled() {
		t.Fatalf("expected auto_path default to true")
	}
}

func TestParseValidationRejectsBadCustomType(t *testing.T) {
	input := []byte(`
version = 1

[[tools]]
name = "demo"
source = "custom"
version = "1.2.3"
url = "https://example.com/demo"
type = "weird"
`)

	_, err := Parse(input)
	if err == nil {
		t.Fatal("expected validation error but got nil")
	}
}
