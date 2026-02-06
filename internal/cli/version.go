package cli

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

var (
	Version   = "0.1.0-dev"
	Commit    = "none"
	BuildTime = "unknown"
	GoVersion = runtime.Version()
)

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
	return fmt.Sprintf("anyisland %s\nCommit: %s\nBuilt: %s\nPlatform: %s/%s\nCompiler: %s", 
		Version, commit, BuildTime, runtime.GOOS, runtime.GOARCH, GoVersion)
}
