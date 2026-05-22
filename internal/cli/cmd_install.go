package cli

import "errors"

func (a *App) cmdInstall(args []string) error {
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

	selected := a.selectedSet(args)
	return rt.Mgr.Sync(backgroundContext(), cfg, selected, a.DryRun, a.Stdout)
}
