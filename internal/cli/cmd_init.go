package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/robert/toolbox/internal/config"
)

func (a *App) cmdInit(args []string) error {
	flags := flag.NewFlagSet("init", flag.ContinueOnError)
	flags.SetOutput(a.Stderr)

	var force bool
	flags.BoolVar(&force, "force", false, "Overwrite existing config file")
	if err := flags.Parse(args); err != nil {
		return err
	}

	if !force {
		if _, err := os.Stat(a.ConfigPath); err == nil {
			return fmt.Errorf("%s already exists (use init --force to overwrite)", a.ConfigPath)
		}
	}

	if a.DryRun {
		fmt.Fprintf(a.Stdout, "would write starter config to %s\n", a.ConfigPath)
		return nil
	}
	if err := os.WriteFile(a.ConfigPath, []byte(config.StarterTOML()), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", a.ConfigPath, err)
	}
	fmt.Fprintf(a.Stdout, "created %s\n", a.ConfigPath)
	return nil
}
