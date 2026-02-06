package cli

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"time"
)

var (
	Version   = "1.0.0"
	Commit    = "none"
	BuildTime = "unknown"
	GoVersion = runtime.Version()
)

func GetEffectiveBuildTime() string {
	if BuildTime != "unknown" {
		return BuildTime
	}
	// Fallback: Get the executable's mod time as a heuristic for build time
	exe, err := os.Executable()
	if err == nil {
		info, err := os.Stat(exe)
		if err == nil {
			return info.ModTime().Format(time.RFC3339) + " (approx)"
		}
	}
	return "unknown"
}

func GetEffectiveCommit() string {
	if Commit != "none" {
		return Commit
	}

	// Use Go's built-in build info (VCS stamping)
	if info, ok := debug.ReadBuildInfo(); ok {
		var rev string
		var dirty bool
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				rev = setting.Value
			case "vcs.modified":
				dirty = setting.Value == "true"
			}
		}
		if rev != "" {
			if dirty {
				return rev + " (dirty)"
			}
			return rev
		}
	}

	return "none"
}

func ShortCommit(commit string) string {
	if len(commit) > 7 {
		return commit[:7]
	}
	return commit
}

func VersionString() string {
	commit := GetEffectiveCommit()
	buildTime := GetEffectiveBuildTime()
	
	return fmt.Sprintf("anyisland %s\nCommit: %s\nBuilt: %s\nPlatform: %s/%s\nCompiler: %s", 
		Version, commit, buildTime, runtime.GOOS, runtime.GOARCH, GoVersion)
}
