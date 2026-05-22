package cli

import (
	"errors"

	"github.com/robert/toolbox/internal/shell"
)

func (a *App) cmdSync(args []string) error {
	if len(args) != 0 {
		return errors.New("sync does not accept positional arguments")
	}

	cfg, err := a.parseConfig(true)
	if err != nil {
		return err
	}
	if len(cfg.Tools) == 0 {
		return errors.New("config has no [[tools]] entries")
	}

	rt, err := a.loadRuntime(cfg)
	if err != nil {
		return err
	}

	if cfg.Defaults.AutoPathEnabled() {
		changed, err := shell.EnsurePathSetup(rt.Dirs.BinDir, a.DryRun)
		if err != nil {
			return err
		}
		for _, path := range changed {
			if a.DryRun {
				_, _ = a.Stdout.Write([]byte("would update " + path + "\n"))
			} else {
				_, _ = a.Stdout.Write([]byte("updated " + path + "\n"))
			}
		}
	}

	return rt.Mgr.Sync(backgroundContext(), cfg, nil, a.DryRun, a.Stdout)
}
