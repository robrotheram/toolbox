# toolbox

A minimal, non-root developer tooling manager for Linux and macOS.

Configure which tools you need (kubectl, helm, k9s, etc.) in a single TOML file. `toolbox` downloads, verifies, and installs them to `~/.local/bin` — no sudo, no nix, no shell scripts per OS.

## Features

- TOML-based config — one file describes your full dev toolset
- Built-in catalog for common tools (kubectl, helm, k9s, k6, sops)
- Custom tools via direct binary or archive (`.tar.gz`, `.zip`) download URL
- `{{os}}`, `{{arch}}`, `{{version}}` placeholders in custom URLs
- Optional SHA-256 checksum verification
- Non-root install to `~/.local/bin` (or any directory you choose)
- Auto-updates `~/.bashrc` and `~/.zshrc` to add the install dir to PATH
- Tracks installed state so installs are idempotent
- `doctor` command to diagnose PATH, directory, and drift issues

## Requirements

- Go 1.22+ to build
- Linux or macOS (amd64 or arm64)

## Installation

```sh
git clone https://github.com/robert/toolbox
cd toolbox
go build -ldflags "-X github.com/robert/toolbox/internal/version.Version=$(git describe --tags --always --dirty)" \
  -o ~/.local/bin/toolbox ./cmd/toolbox
```

Or run directly without installing:

```sh
go run ./cmd/toolbox [flags] <command>
```

> **Version fallback**: if no `-ldflags` are passed and no git tag is present, `toolbox version` prints `dev`.

## Quick start

```sh
# Generate a starter config in the current directory
toolbox init

# Edit toolbox.toml, then install all tools
toolbox sync

# Reload your shell (or source the rc file) to pick up PATH changes
source ~/.bashrc   # or ~/.zshrc
```

## Config format

```toml
version = 1

[defaults]
bin_dir   = "~/.local/bin"              # where binaries are installed
state_dir = "~/.local/share/toolbox"
cache_dir = "~/.cache/toolbox"
auto_path = true                        # write PATH line to bash/zsh rc files

# Built-in tools — just name and version
[[tools]]
name    = "kubectl"
source  = "builtin"
version = "1.30.2"

[[tools]]
name    = "helm"
source  = "builtin"
version = "3.15.2"

[[tools]]
name    = "k9s"
source  = "builtin"
version = "0.32.5"

# Custom tool — direct binary download
[[tools]]
name    = "my-tool"
source  = "custom"
type    = "binary"
version = "2.1.0"
url     = "https://example.com/my-tool-{{os}}-{{arch}}-{{version}}"
# sha256 = "optional-checksum"

# Custom tool — archive download
[[tools]]
name           = "another-tool"
source         = "custom"
type           = "archive"
version        = "1.0.0"
url            = "https://example.com/another-tool-{{os}}-{{arch}}-{{version}}.tar.gz"
archive_binary = "another-tool"        # path inside the archive to extract
binary_name    = "another-tool"        # name to install as (defaults to name)
```

### URL placeholders

| Placeholder          | Expands to                    |
|----------------------|-------------------------------|
| `{{os}}`             | `linux` or `darwin`           |
| `{{arch}}`           | `amd64` or `arm64`            |
| `{{version}}`        | version without leading `v`   |
| `{{version_with_v}}` | version with leading `v`      |

### Built-in catalog

| Tool    | Default version |
|---------|----------------|
| kubectl | 1.30.2          |
| helm    | 3.15.2          |
| k9s     | 0.32.5          |
| k6      | 0.52.0          |
| sops    | 3.9.0           |
| sops    | 3.9.0           |

## Commands

```
toolbox [global flags] <command> [args]

Commands:
  init      Create a starter toolbox.toml in the current directory
  sync      Install or update all tools defined in config
  install   Install specific tool(s) by name  (e.g. install kubectl helm)
  list      List all managed tools and their install paths
  remove    Remove one or more managed tools   (e.g. remove kubectl)
  doctor    Check PATH, install directory, and tool state for problems
  version   Print the toolbox version

Global flags:
  -config <path>   Path to TOML config file (default: toolbox.toml)
  -bin-dir <path>  Override the binary install directory
  -dry-run         Print planned actions without making any changes
  -verbose         Verbose output
```

## Directory layout

```
~/.local/bin/                         # binaries (on PATH)
~/.local/share/toolbox/state.json # installed-state manifest
~/.cache/toolbox/                 # temporary download artifacts
```

## How PATH setup works

When `auto_path = true` (the default), running `sync` will append the following block to both `~/.bashrc` and `~/.zshrc` if it is not already present. The block is idempotent — repeated runs will not add duplicate entries. A backup of each rc file is made before any edit.

```sh
# >>> toolbox >>>
export PATH="$HOME/.local/bin:$PATH"
# <<< toolbox <<<
```

To disable automatic rc-file modification, set `auto_path = false` in `[defaults]` and add the export line yourself.

## Release automation

This repository uses `.github/workflows/release.yml` to create GitHub releases automatically when changes are merged to `main`.

- Trigger: every push to `main` (including merge commits)
- Versioning: semver tags (`vMAJOR.MINOR.PATCH`) are kept and auto-bumped by patch
- First release tag: `v0.1.0` if no semver tag exists yet
- Artifacts: currently builds only `linux/amd64` (`toolbox-linux-amd64`)
- Extensibility: add more targets by extending the workflow matrix

## Building and testing

```sh
go build ./...
go test ./...
```
