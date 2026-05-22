package cli

import (
	"fmt"

	"github.com/robert/toolbox/internal/version"
)

func (a *App) cmdVersion(_ []string) error {
	fmt.Fprintln(a.Stdout, version.Get())
	return nil
}
