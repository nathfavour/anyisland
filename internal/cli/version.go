package cli

import (
	"fmt"
	"runtime"
)

var (
	Version   = "0.1.0-dev"
	Commit    = "none"
	BuildTime = "unknown"
	GoVersion = runtime.Version()
)

func VersionString() string {
	return fmt.Sprintf("anyisland %s\nCommit: %s\nBuilt: %s\nPlatform: %s/%s\nCompiler: %s", 
		Version, Commit, BuildTime, runtime.GOOS, runtime.GOARCH, GoVersion)
}
