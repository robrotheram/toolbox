package install

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Request struct {
	Name            string
	Version         string
	BinaryName      string
	DownloadURL     string
	PackageType     string
	ArchiveBinaries []string
	SHA256          string
	DestDir         string
	CacheDir        string
	DryRun          bool
}

type Result struct {
	Name         string
	Version      string
	InstallPath  string
	InstallPaths []string
	SHA256       string
}

func Install(ctx context.Context, req Request) (Result, error) {
	if strings.TrimSpace(req.BinaryName) == "" {
		return Result{}, fmt.Errorf("binary name is required")
	}
	if strings.TrimSpace(req.DownloadURL) == "" {
		return Result{}, fmt.Errorf("download URL is required")
	}
	if strings.TrimSpace(req.DestDir) == "" {
		return Result{}, fmt.Errorf("destination directory is required")
	}
	if strings.TrimSpace(req.CacheDir) == "" {
		return Result{}, fmt.Errorf("cache directory is required")
	}

	destPath := filepath.Join(req.DestDir, req.BinaryName)
	if req.DryRun {
		var installPaths []string
		if len(req.ArchiveBinaries) > 0 {
			for _, binName := range req.ArchiveBinaries {
				installPaths = append(installPaths, filepath.Join(req.DestDir, filepath.Base(binName)))
			}
		}
		return Result{
			Name:         req.Name,
			Version:      req.Version,
			InstallPath:  destPath,
			InstallPaths: installPaths,
		}, nil
	}

	if err := os.MkdirAll(req.DestDir, 0o755); err != nil {
		return Result{}, fmt.Errorf("create destination directory: %w", err)
	}
	if err := os.MkdirAll(req.CacheDir, 0o755); err != nil {
		return Result{}, fmt.Errorf("create cache directory: %w", err)
	}

	tmpDir, err := os.MkdirTemp(req.CacheDir, "toolbox-*")
	if err != nil {
		return Result{}, fmt.Errorf("create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	artifactName := "artifact"
	if idx := strings.LastIndex(req.DownloadURL, "/"); idx >= 0 && idx < len(req.DownloadURL)-1 {
		artifactName = req.DownloadURL[idx+1:]
	}
	artifactPath := filepath.Join(tmpDir, artifactName)
	if err := Download(ctx, req.DownloadURL, artifactPath); err != nil {
		return Result{}, err
	}
	if err := VerifySHA256(artifactPath, req.SHA256); err != nil {
		return Result{}, err
	}

	var installPaths []string
	switch req.PackageType {
	case "binary":
		stagingPath := destPath + ".tmp"
		if err := copyExecutable(artifactPath, stagingPath); err != nil {
			return Result{}, err
		}
		if err := os.Rename(stagingPath, destPath); err != nil {
			return Result{}, fmt.Errorf("move %s -> %s: %w", stagingPath, destPath, err)
		}
	case "archive":
		if len(req.ArchiveBinaries) == 0 {
			return Result{}, fmt.Errorf("archive_binaries is required for archive package type")
		}
		for _, binName := range req.ArchiveBinaries {
			destBinPath := filepath.Join(req.DestDir, filepath.Base(binName))
			stagingPath := destBinPath + ".tmp"
			if err := ExtractBinary(artifactPath, binName, stagingPath); err != nil {
				return Result{}, fmt.Errorf("extract %s: %w", binName, err)
			}
			if err := os.Rename(stagingPath, destBinPath); err != nil {
				return Result{}, fmt.Errorf("move %s -> %s: %w", stagingPath, destBinPath, err)
			}
			installPaths = append(installPaths, destBinPath)
		}
	default:
		return Result{}, fmt.Errorf("unsupported package type %q", req.PackageType)
	}

	sum, err := FileSHA256(destPath)
	if err != nil {
		return Result{}, err
	}

	return Result{
		Name:         req.Name,
		Version:      req.Version,
		InstallPath:  destPath,
		InstallPaths: installPaths,
		SHA256:       sum,
	}, nil
}

func copyExecutable(sourcePath, destinationPath string) error {
	in, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("open %s: %w", sourcePath, err)
	}
	defer in.Close()

	out, err := os.OpenFile(destinationPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return fmt.Errorf("create %s: %w", destinationPath, err)
	}

	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return fmt.Errorf("copy %s -> %s: %w", sourcePath, destinationPath, err)
	}

	if err := out.Close(); err != nil {
		return fmt.Errorf("close %s: %w", destinationPath, err)
	}
	if err := os.Chmod(destinationPath, 0o755); err != nil {
		return fmt.Errorf("chmod %s: %w", destinationPath, err)
	}
	return nil
}
