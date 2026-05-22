package cli

import "errors"

func (a *App) cmdList(args []string) error {
	if len(args) != 0 {
		return errors.New("list does not accept positional arguments")
	}

	cfg, err := a.parseConfig(false)
	if err != nil {
		return err
	}

	rt, err := a.loadRuntime(cfg)
	if err != nil {
		return err
	}

	manifest, err := rt.Store.Load()
	if err != nil {
		return err
	}
	a.printStateEntries(manifest.Sorted())
	return nil
}
