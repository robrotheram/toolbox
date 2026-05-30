package registry

import (
	"fmt"
	"sort"
	"strings"

	"github.com/robert/toolbox/internal/config"
	"github.com/robert/toolbox/internal/platform"
)

type ResolvedTool struct {
	Name            string
	Version         string
	BinaryName      string
	DownloadURL     string
	PackageType     string
	ArchiveBinaries []string
}

type builtinDef struct {
	DefaultVersion string
	Resolve        func(version string, target platform.Target) ResolvedTool
}

var builtins = map[string]builtinDef{
	"kubectl": {
		DefaultVersion: "1.30.2",
		Resolve: func(version string, target platform.Target) ResolvedTool {
			v := trimV(version)
			return ResolvedTool{
				Name:        "kubectl",
				Version:     v,
				BinaryName:  "kubectl",
				DownloadURL: fmt.Sprintf("https://dl.k8s.io/release/v%s/bin/%s/%s/kubectl", v, target.OS, target.Arch),
				PackageType: config.ToolTypeBinary,
			}
		},
	},
	"helm": {
		DefaultVersion: "3.15.2",
		Resolve: func(version string, target platform.Target) ResolvedTool {
			v := trimV(version)
			return ResolvedTool{
				Name:            "helm",
				Version:         v,
				BinaryName:      "helm",
				DownloadURL:     fmt.Sprintf("https://get.helm.sh/helm-v%s-%s-%s.tar.gz", v, target.OS, target.Arch),
				PackageType:     config.ToolTypeArchive,
				ArchiveBinaries: []string{fmt.Sprintf("%s-%s/helm", target.OS, target.Arch)},
			}
		},
	},
	"k9s": {
		DefaultVersion: "0.32.5",
		Resolve: func(version string, target platform.Target) ResolvedTool {
			v := trimV(version)
			return ResolvedTool{
				Name:            "k9s",
				Version:         v,
				BinaryName:      "k9s",
				DownloadURL:     fmt.Sprintf("https://github.com/derailed/k9s/releases/download/v%s/k9s_%s_%s.tar.gz", v, toTitleOS(target.OS), target.Arch),
				PackageType:     config.ToolTypeArchive,
				ArchiveBinaries: []string{"k9s"},
			}
		},
	},
	"k6": {
		DefaultVersion: "0.52.0",
		Resolve: func(version string, target platform.Target) ResolvedTool {
			v := trimV(version)
			return ResolvedTool{
				Name:            "k6",
				Version:         v,
				BinaryName:      "k6",
				DownloadURL:     fmt.Sprintf("https://github.com/grafana/k6/releases/download/v%s/k6-v%s-%s-%s.tar.gz", v, v, target.OS, target.Arch),
				PackageType:     config.ToolTypeArchive,
				ArchiveBinaries: []string{fmt.Sprintf("k6-v%s-%s-%s/k6", v, target.OS, target.Arch)},
			}
		},
	},
	"sops": {
		DefaultVersion: "3.9.0",
		Resolve: func(version string, target platform.Target) ResolvedTool {
			v := trimV(version)
			return ResolvedTool{
				Name:        "sops",
				Version:     v,
				BinaryName:  "sops",
				DownloadURL: fmt.Sprintf("https://github.com/getsops/sops/releases/download/v%s/sops-v%s.%s.%s", v, v, target.OS, target.Arch),
				PackageType: config.ToolTypeBinary,
			}
		},
	},
	"secretspec": {
		DefaultVersion: "0.11.0",
		Resolve: func(version string, target platform.Target) ResolvedTool {
			v := trimV(version)
			triple := toSecretspecTarget(target)
			return ResolvedTool{
				Name:            "secretspec",
				Version:         v,
				BinaryName:      "secretspec",
				DownloadURL:     fmt.Sprintf("https://github.com/cachix/secretspec/releases/download/v%s/secretspec-%s.tar.xz", v, triple),
				PackageType:     config.ToolTypeArchive,
				ArchiveBinaries: []string{"secretspec"},
			}
		},
	},
}

func ResolveBuiltin(name, version string, target platform.Target) (ResolvedTool, error) {
	key := strings.ToLower(strings.TrimSpace(name))
	def, ok := builtins[key]
	if !ok {
		return ResolvedTool{}, fmt.Errorf("unsupported built-in tool %q (supported: %s)", name, strings.Join(KnownTools(), ", "))
	}
	if strings.TrimSpace(version) == "" {
		version = def.DefaultVersion
	}
	resolved := def.Resolve(version, target)
	return resolved, nil
}

func KnownTools() []string {
	tools := make([]string, 0, len(builtins))
	for k := range builtins {
		tools = append(tools, k)
	}
	sort.Strings(tools)
	return tools
}

func trimV(version string) string {
	return strings.TrimPrefix(strings.TrimSpace(version), "v")
}

func toTitleOS(goos string) string {
	switch goos {
	case "linux":
		return "Linux"
	case "darwin":
		return "Darwin"
	default:
		return goos
	}
}

func toSecretspecTarget(target platform.Target) string {
	osPart := target.OS
	archPart := target.Arch

	switch archPart {
	case "amd64":
		archPart = "x86_64"
	case "arm64":
		archPart = "aarch64"
	}

	switch osPart {
	case "linux":
		osPart = "unknown-linux-gnu"
	case "darwin":
		osPart = "apple-darwin"
	}

	return fmt.Sprintf("%s-%s", archPart, osPart)
}
