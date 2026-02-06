package cli

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

var (
	Version   = "0.1.0-dev"
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
	// Fallback to local git if in dev
	cmd := exec.Command("git", "rev-parse", "HEAD")
	out, err := cmd.Output()
	if err == nil {
		return strings.TrimSpace(string(out))
	}
	return "none"
}

func VersionString() string {
	commit := GetEffectiveCommit()
	if Commit == "none" && commit != "none" {
		commit += " (local)"
	}
	buildTime := GetEffectiveBuildTime()
	
	return fmt.Sprintf("anyisland %s\nCommit: %s\nBuilt: %s\nPlatform: %s/%s\nCompiler: %s", 
		Version, commit, buildTime, runtime.GOOS, runtime.GOARCH, GoVersion)
}
