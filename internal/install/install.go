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
	Name          string
	Version       string
	BinaryName    string
	DownloadURL   string
	PackageType   string
	ArchiveBinary string
	SHA256        string
	DestDir       string
	CacheDir      string
	DryRun        bool
}

type Result struct {
	Name        string
	Version     string
	InstallPath string
	SHA256      string
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
		return Result{
			Name:        req.Name,
			Version:     req.Version,
			InstallPath: destPath,
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

	stagingPath := destPath + ".tmp"
	switch req.PackageType {
	case "binary":
		if err := copyExecutable(artifactPath, stagingPath); err != nil {
			return Result{}, err
		}
	case "archive":
		if strings.TrimSpace(req.ArchiveBinary) == "" {
			return Result{}, fmt.Errorf("archive binary is required for archive package type")
		}
		if err := ExtractBinary(artifactPath, req.ArchiveBinary, stagingPath); err != nil {
			return Result{}, err
		}
	default:
		return Result{}, fmt.Errorf("unsupported package type %q", req.PackageType)
	}

	if err := os.Rename(stagingPath, destPath); err != nil {
		return Result{}, fmt.Errorf("move %s -> %s: %w", stagingPath, destPath, err)
	}

	sum, err := FileSHA256(destPath)
	if err != nil {
		return Result{}, err
	}

	return Result{
		Name:        req.Name,
		Version:     req.Version,
		InstallPath: destPath,
		SHA256:      sum,
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
