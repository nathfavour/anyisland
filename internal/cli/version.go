package cli

import (
	"fmt"
	"runtime"
)

var (
	Version   = "0.1.0-dev"
	Commit    = "none"
	BuildTime = "unknown"
)

func VersionString() string {
	return fmt.Sprintf("anyisland %s (%s) %s/%s build:%s", Version, Commit, runtime.GOOS, runtime.GOARCH, BuildTime)
}
