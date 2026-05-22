package config

import "strings"

const (
	SourceBuiltin = "builtin"
	SourceCustom  = "custom"

	ToolTypeBinary  = "binary"
	ToolTypeArchive = "archive"
)

type Config struct {
	Version  int      `toml:"version"`
	Defaults Defaults `toml:"defaults"`
	Tools    []Tool   `toml:"tools"`
}

type Defaults struct {
	BinDir   string `toml:"bin_dir"`
	StateDir string `toml:"state_dir"`
	CacheDir string `toml:"cache_dir"`
	AutoPath *bool  `toml:"auto_path"`
}

func (d Defaults) AutoPathEnabled() bool {
	if d.AutoPath == nil {
		return true
	}
	return *d.AutoPath
}

type Tool struct {
	Name            string `toml:"name"`
	Source          string `toml:"source"`
	Version         string `toml:"version"`
	URL             string `toml:"url"`
	Type            string `toml:"type"`
	ArchiveBinary   string `toml:"archive_binary"`
	BinaryName      string `toml:"binary_name"`
	SHA256          string `toml:"sha256"`
	StripComponents int    `toml:"strip_components"`
}

func StarterTOML() string {
	return strings.TrimSpace(`
version = 1

[defaults]
bin_dir = "~/.local/bin"
state_dir = "~/.local/share/toolbox"
cache_dir = "~/.cache/toolbox"
auto_path = true

[[tools]]
name = "kubectl"
source = "builtin"
version = "1.30.2"

[[tools]]
name = "helm"
source = "builtin"
version = "3.15.2"

[[tools]]
name = "my-custom-tool"
source = "custom"
type = "archive"
version = "1.2.3"
url = "https://example.com/my-tool-{{os}}-{{arch}}-{{version}}.tar.gz"
archive_binary = "my-tool"
binary_name = "my-tool"
# sha256 = "optional-checksum-here"
`) + "\n"
}
