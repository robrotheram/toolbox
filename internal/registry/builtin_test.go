package registry

import (
	"strings"
	"testing"

	"github.com/robert/toolbox/internal/platform"
)

func TestResolveBuiltinHelm(t *testing.T) {
	target := platform.Target{OS: "linux", Arch: "amd64"}
	resolved, err := ResolveBuiltin("helm", "3.15.2", target)
	if err != nil {
		t.Fatalf("ResolveBuiltin() error: %v", err)
	}
	if resolved.PackageType != "archive" {
		t.Fatalf("expected archive package type, got %q", resolved.PackageType)
	}
	if !strings.Contains(resolved.DownloadURL, "helm-v3.15.2-linux-amd64.tar.gz") {
		t.Fatalf("unexpected helm URL %q", resolved.DownloadURL)
	}
}
