package version

import (
	"runtime/debug"
)

// Version is set at build time via:
//
//	go build -ldflags "-X github.com/robert/toolbox/internal/version.Version=v1.2.3"
var Version = ""

func Get() string {
	if Version != "" {
		return Version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return "dev"
}
