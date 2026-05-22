package cli

import (
	"errors"
	"fmt"

	"github.com/robert/toolbox/internal/doctor"
)

func (a *App) cmdDoctor(args []string) error {
	if len(args) != 0 {
		return errors.New("doctor does not accept positional arguments")
	}

	cfg, err := a.parseConfig(false)
	if err != nil {
		return err
	}

	rt, err := a.loadRuntime(cfg)
	if err != nil {
		return err
	}

	report, err := doctor.Run(a.ConfigPath, rt.Dirs.BinDir, rt.Store)
	if err != nil {
		return err
	}

	for _, info := range report.Info {
		fmt.Fprintf(a.Stdout, "INFO: %s\n", info)
	}
	for _, warning := range report.Warnings {
		fmt.Fprintf(a.Stdout, "WARN: %s\n", warning)
	}
	for _, failure := range report.Errors {
		fmt.Fprintf(a.Stdout, "ERROR: %s\n", failure)
	}
	if len(report.Errors) == 0 {
		fmt.Fprintln(a.Stdout, "doctor checks passed")
		return nil
	}
	return fmt.Errorf("doctor found %d error(s)", len(report.Errors))
}
