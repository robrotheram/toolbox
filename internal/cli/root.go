package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/robert/toolbox/internal/config"
	"github.com/robert/toolbox/internal/paths"
	"github.com/robert/toolbox/internal/platform"
	"github.com/robert/toolbox/internal/reconcile"
	"github.com/robert/toolbox/internal/state"
)

type App struct {
	Stdout io.Writer
	Stderr io.Writer

	ConfigPath     string
	BinDirOverride string
	DryRun         bool
	Verbose        bool
}

func Execute(args []string, stdout, stderr io.Writer) int {
	app := &App{
		Stdout: stdout,
		Stderr: stderr,
	}
	if err := app.Run(args); err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	return 0
}

func (a *App) Run(args []string) error {
	flags := flag.NewFlagSet("toolbox", flag.ContinueOnError)
	flags.SetOutput(a.Stderr)

	var help bool
	flags.BoolVar(&help, "help", false, "Show help")
	flags.BoolVar(&help, "h", false, "Show help")
	flags.StringVar(&a.ConfigPath, "config", "toolbox.toml", "Path to TOML config")
	flags.StringVar(&a.BinDirOverride, "bin-dir", "", "Override binary install directory")
	flags.BoolVar(&a.DryRun, "dry-run", false, "Print actions without mutating state")
	flags.BoolVar(&a.Verbose, "verbose", false, "Verbose output")

	if err := flags.Parse(args); err != nil {
		return err
	}
	if help {
		a.printUsage()
		return nil
	}

	rest := flags.Args()
	if len(rest) == 0 {
		a.printUsage()
		return errors.New("command is required")
	}

	command := strings.ToLower(rest[0])
	commandArgs := rest[1:]

	switch command {
	case "init":
		return a.cmdInit(commandArgs)
	case "sync":
		return a.cmdSync(commandArgs)
	case "install":
		return a.cmdInstall(commandArgs)
	case "list":
		return a.cmdList(commandArgs)
	case "remove":
		return a.cmdRemove(commandArgs)
	case "doctor":
		return a.cmdDoctor(commandArgs)
	case "version":
		return a.cmdVersion(commandArgs)
	case "help":
		a.printUsage()
		return nil
	default:
		return fmt.Errorf("unknown command %q", command)
	}
}

type runtimeContext struct {
	Config config.Config
	Dirs   paths.Dirs
	Target platform.Target
	Store  state.Store
	Mgr    reconcile.Manager
}

func (a *App) loadRuntime(cfg config.Config) (runtimeContext, error) {
	dirs, err := paths.Resolve(cfg.Defaults, a.BinDirOverride)
	if err != nil {
		return runtimeContext{}, err
	}
	if !a.DryRun {
		if err := paths.Ensure(dirs); err != nil {
			return runtimeContext{}, err
		}
	}

	target, err := platform.Detect()
	if err != nil {
		return runtimeContext{}, err
	}

	store := state.New(dirs.StateDir)
	return runtimeContext{
		Config: cfg,
		Dirs:   dirs,
		Target: target,
		Store:  store,
		Mgr:    reconcile.NewManager(dirs, target, store),
	}, nil
}

func (a *App) parseConfig(required bool) (config.Config, error) {
	cfg, err := config.ParseFile(a.ConfigPath)
	if err == nil {
		return cfg, nil
	}
	if !required && os.IsNotExist(err) {
		return config.Config{Version: 1}, nil
	}
	return config.Config{}, err
}

func (a *App) selectedSet(names []string) map[string]struct{} {
	if len(names) == 0 {
		return nil
	}
	selected := map[string]struct{}{}
	for _, name := range names {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			continue
		}
		selected[trimmed] = struct{}{}
	}
	return selected
}

func (a *App) printStateEntries(entries []state.Entry) {
	if len(entries) == 0 {
		fmt.Fprintln(a.Stdout, "No managed tools installed.")
		return
	}
	fmt.Fprintln(a.Stdout, "NAME\tVERSION\tSOURCE\tPATH")
	for _, entry := range entries {
		fmt.Fprintf(a.Stdout, "%s\t%s\t%s\t%s\n", entry.Name, entry.Version, entry.Source, entry.InstallPath)
	}
}

func (a *App) printUsage() {
	fmt.Fprintln(a.Stdout, "toolbox - non-root developer tool manager")
	fmt.Fprintln(a.Stdout, "")
	fmt.Fprintln(a.Stdout, "Usage:")
	fmt.Fprintln(a.Stdout, "  toolbox [global flags] <command> [args]")
	fmt.Fprintln(a.Stdout, "")
	fmt.Fprintln(a.Stdout, "Commands:")
	fmt.Fprintln(a.Stdout, "  init      Create starter toolbox.toml")
	fmt.Fprintln(a.Stdout, "  sync      Reconcile all tools from config")
	fmt.Fprintln(a.Stdout, "  install   Install all tools or selected tool names")
	fmt.Fprintln(a.Stdout, "  list      List managed installed tools")
	fmt.Fprintln(a.Stdout, "  remove    Remove managed tool(s)")
	fmt.Fprintln(a.Stdout, "  doctor    Run environment diagnostics")
	fmt.Fprintln(a.Stdout, "  version   Print toolbox version")
	fmt.Fprintln(a.Stdout, "")
	fmt.Fprintln(a.Stdout, "Global flags:")
	fmt.Fprintln(a.Stdout, "  -config <path>   Path to toolbox TOML config (default: toolbox.toml)")
	fmt.Fprintln(a.Stdout, "  -bin-dir <path>  Override install directory (default from config or ~/.local/bin)")
	fmt.Fprintln(a.Stdout, "  -dry-run         Print actions without changes")
	fmt.Fprintln(a.Stdout, "  -verbose         Verbose output")
}

func sortKeys(set map[string]struct{}) []string {
	keys := make([]string, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func backgroundContext() context.Context {
	return context.Background()
}
