package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	blockStart = "# >>> toolbox >>>"
	blockEnd   = "# <<< toolbox <<<"
)

func EnsurePathSetup(binDir string, dryRun bool) ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("resolve user home: %w", err)
	}

	targets := []string{
		filepath.Join(home, ".bashrc"),
		filepath.Join(home, ".zshrc"),
	}

	pathExpression := shellPathExpression(binDir, home)
	var changed []string
	for _, target := range targets {
		updated, err := ensureFile(target, pathExpression, dryRun)
		if err != nil {
			return nil, err
		}
		if updated {
			changed = append(changed, target)
		}
	}
	return changed, nil
}

func ensureFile(path, pathExpression string, dryRun bool) (bool, error) {
	current, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return false, fmt.Errorf("read %s: %w", path, err)
	}

	next := injectPathBlock(string(current), pathExpression)
	if next == string(current) {
		return false, nil
	}
	if dryRun {
		return true, nil
	}

	if len(current) > 0 {
		backupPath := fmt.Sprintf("%s.bak.%d", path, time.Now().Unix())
		if err := os.WriteFile(backupPath, current, 0o644); err != nil {
			return false, fmt.Errorf("write backup %s: %w", backupPath, err)
		}
	}

	if err := os.WriteFile(path, []byte(next), 0o644); err != nil {
		return false, fmt.Errorf("write %s: %w", path, err)
	}
	return true, nil
}

func injectPathBlock(content, pathExpression string) string {
	block := fmt.Sprintf("%s\nexport PATH=\"%s:$PATH\"\n%s", blockStart, pathExpression, blockEnd)

	start := strings.Index(content, blockStart)
	end := strings.Index(content, blockEnd)
	if start >= 0 && end > start {
		end += len(blockEnd)
		return content[:start] + block + content[end:]
	}

	trimmed := strings.TrimRight(content, "\n")
	if trimmed == "" {
		return block + "\n"
	}
	return trimmed + "\n\n" + block + "\n"
}

func shellPathExpression(binDir, home string) string {
	if binDir == home {
		return "$HOME"
	}
	if strings.HasPrefix(binDir, home+"/") {
		return strings.Replace(binDir, home, "$HOME", 1)
	}
	return binDir
}
