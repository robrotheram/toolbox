package platform

import (
	"fmt"
	"runtime"
)

type Target struct {
	OS   string
	Arch string
}

func Detect() (Target, error) {
	target := Target{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}

	switch target.OS {
	case "linux", "darwin":
	default:
		return Target{}, fmt.Errorf("unsupported OS: %s (supported: linux, darwin)", target.OS)
	}

	switch target.Arch {
	case "amd64", "arm64":
	default:
		return Target{}, fmt.Errorf("unsupported architecture: %s (supported: amd64, arm64)", target.Arch)
	}

	return target, nil
}
