package install

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ulikunitz/xz"
)

func ExtractBinary(archivePath, archiveBinary, destination string) error {
	switch {
	case strings.HasSuffix(archivePath, ".zip"):
		return extractFromZip(archivePath, archiveBinary, destination)
	case strings.HasSuffix(archivePath, ".tar.gz"), strings.HasSuffix(archivePath, ".tgz"):
		return extractFromTarGz(archivePath, archiveBinary, destination)
	case strings.HasSuffix(archivePath, ".tar.xz"):
		return extractFromTarXz(archivePath, archiveBinary, destination)
	default:
		return fmt.Errorf("unsupported archive format for %s", archivePath)
	}
}

func extractFromTarGz(path, wanted, destination string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open archive: %w", err)
	}
	defer file.Close()

	gz, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("open gzip stream: %w", err)
	}
	defer gz.Close()

	reader := tar.NewReader(gz)
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read tar entry: %w", err)
		}

		if header.Typeflag != tar.TypeReg {
			continue
		}

		if !matchesArchivePath(header.Name, wanted) {
			continue
		}

		if err := writeExecutable(destination, reader); err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("binary %q not found in %s", wanted, path)
}

func extractFromTarXz(path, wanted, destination string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open archive: %w", err)
	}
	defer file.Close()

	xzReader, err := xz.NewReader(file)
	if err != nil {
		return fmt.Errorf("open xz stream: %w", err)
	}

	reader := tar.NewReader(xzReader)
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read tar entry: %w", err)
		}

		if header.Typeflag != tar.TypeReg {
			continue
		}

		if !matchesArchivePath(header.Name, wanted) {
			continue
		}

		if err := writeExecutable(destination, reader); err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("binary %q not found in %s", wanted, path)
}

func extractFromZip(path, wanted, destination string) error {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return fmt.Errorf("open zip archive: %w", err)
	}
	defer reader.Close()

	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			continue
		}
		if !matchesArchivePath(file.Name, wanted) {
			continue
		}

		rc, err := file.Open()
		if err != nil {
			return fmt.Errorf("open zip entry %s: %w", file.Name, err)
		}

		writeErr := writeExecutable(destination, rc)
		closeErr := rc.Close()
		if writeErr != nil {
			return writeErr
		}
		if closeErr != nil {
			return fmt.Errorf("close zip entry %s: %w", file.Name, closeErr)
		}
		return nil
	}
	return fmt.Errorf("binary %q not found in %s", wanted, path)
}

func matchesArchivePath(entryName, wanted string) bool {
	entry := filepath.Clean(strings.TrimSpace(entryName))
	needle := filepath.Clean(strings.TrimSpace(wanted))
	if strings.Contains(needle, "/") {
		return entry == needle
	}
	return filepath.Base(entry) == needle
}

func writeExecutable(destination string, source io.Reader) error {
	out, err := os.OpenFile(destination, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return fmt.Errorf("create %s: %w", destination, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, source); err != nil {
		return fmt.Errorf("write %s: %w", destination, err)
	}
	if err := out.Chmod(0o755); err != nil {
		return fmt.Errorf("chmod %s: %w", destination, err)
	}
	return nil
}
