package cli

import (
	"errors"
	"strings"
)

func (a *App) cmdRemove(args []string) error {
	if len(args) == 0 {
		return errors.New("remove requires at least one tool name")
	}

	cfg, err := a.parseConfig(false)
	if err != nil {
		return err
	}

	rt, err := a.loadRuntime(cfg)
	if err != nil {
		return err
	}

	names := make([]string, 0, len(args))
	for _, name := range args {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			continue
		}
		names = append(names, trimmed)
	}
	if len(names) == 0 {
		return errors.New("remove requires at least one non-empty tool name")
	}
	return rt.Mgr.Remove(names, a.DryRun, a.Stdout)
}
