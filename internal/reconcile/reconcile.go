package reconcile

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/robert/toolbox/internal/config"
	"github.com/robert/toolbox/internal/install"
	"github.com/robert/toolbox/internal/paths"
	"github.com/robert/toolbox/internal/platform"
	"github.com/robert/toolbox/internal/registry"
	"github.com/robert/toolbox/internal/state"
)

type Manager struct {
	Dirs   paths.Dirs
	Target platform.Target
	Store  state.Store
}

func NewManager(dirs paths.Dirs, target platform.Target, store state.Store) Manager {
	return Manager{
		Dirs:   dirs,
		Target: target,
		Store:  store,
	}
}

func (m Manager) Sync(ctx context.Context, cfg config.Config, selected map[string]struct{}, dryRun bool, out io.Writer) error {
	manifest, err := m.Store.Load()
	if err != nil {
		return err
	}

	found := map[string]struct{}{}
	for _, tool := range cfg.Tools {
		if selected != nil {
			if _, ok := selected[tool.Name]; !ok {
				continue
			}
		}
		found[tool.Name] = struct{}{}

		req, source, resolvedURL, err := m.installRequest(tool, dryRun)
		if err != nil {
			return fmt.Errorf("prepare %s: %w", tool.Name, err)
		}

		result, err := install.Install(ctx, req)
		if err != nil {
			return fmt.Errorf("install %s: %w", tool.Name, err)
		}

		fmt.Fprintf(out, "installed %s@%s -> %s\n", result.Name, result.Version, result.InstallPath)
		if dryRun {
			continue
		}
		manifest.Upsert(state.Entry{
			Name:        result.Name,
			Version:     result.Version,
			Source:      source,
			URL:         resolvedURL,
			InstallPath: result.InstallPath,
			SHA256:      result.SHA256,
			UpdatedAt:   time.Now().UTC().Format(time.RFC3339),
		})
	}

	if selected != nil {
		var missing []string
		for name := range selected {
			if _, ok := found[name]; !ok {
				missing = append(missing, name)
			}
		}
		if len(missing) > 0 {
			sort.Strings(missing)
			return fmt.Errorf("tools not found in config: %s", strings.Join(missing, ", "))
		}
	}

	if dryRun {
		return nil
	}
	return m.Store.Save(manifest)
}

func (m Manager) Remove(names []string, dryRun bool, out io.Writer) error {
	manifest, err := m.Store.Load()
	if err != nil {
		return err
	}

	for _, name := range names {
		entry, ok := manifest.Remove(name)
		if !ok {
			return fmt.Errorf("tool %q is not currently managed", name)
		}
		if !dryRun {
			if err := os.Remove(entry.InstallPath); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("remove %s: %w", entry.InstallPath, err)
			}
		}
		fmt.Fprintf(out, "removed %s\n", name)
	}

	if dryRun {
		return nil
	}
	return m.Store.Save(manifest)
}

func (m Manager) installRequest(tool config.Tool, dryRun bool) (install.Request, string, string, error) {
	switch tool.Source {
	case config.SourceBuiltin:
		resolved, err := registry.ResolveBuiltin(tool.Name, tool.Version, m.Target)
		if err != nil {
			return install.Request{}, "", "", err
		}
		req := install.Request{
			Name:          tool.Name,
			Version:       resolved.Version,
			BinaryName:    resolved.BinaryName,
			DownloadURL:   resolved.DownloadURL,
			PackageType:   resolved.PackageType,
			ArchiveBinary: resolved.ArchiveBinary,
			SHA256:        tool.SHA256,
			DestDir:       m.Dirs.BinDir,
			CacheDir:      m.Dirs.CacheDir,
			DryRun:        dryRun,
		}
		return req, config.SourceBuiltin, resolved.DownloadURL, nil
	case config.SourceCustom:
		url := interpolate(tool.URL, m.Target, tool.Version)
		binaryName := tool.BinaryName
		if binaryName == "" {
			binaryName = tool.Name
		}
		archiveBinary := interpolate(tool.ArchiveBinary, m.Target, tool.Version)
		return install.Request{
			Name:          tool.Name,
			Version:       tool.Version,
			BinaryName:    filepath.Base(binaryName),
			DownloadURL:   url,
			PackageType:   tool.Type,
			ArchiveBinary: archiveBinary,
			SHA256:        tool.SHA256,
			DestDir:       m.Dirs.BinDir,
			CacheDir:      m.Dirs.CacheDir,
			DryRun:        dryRun,
		}, config.SourceCustom, url, nil
	default:
		return install.Request{}, "", "", fmt.Errorf("unknown source %q", tool.Source)
	}
}

func interpolate(value string, target platform.Target, version string) string {
	replacer := strings.NewReplacer(
		"{{os}}", target.OS,
		"{{arch}}", target.Arch,
		"{{version}}", strings.TrimPrefix(version, "v"),
		"{{version_with_v}}", ensureV(version),
	)
	return replacer.Replace(value)
}

func ensureV(version string) string {
	v := strings.TrimSpace(version)
	if strings.HasPrefix(v, "v") {
		return v
	}
	return "v" + v
}
