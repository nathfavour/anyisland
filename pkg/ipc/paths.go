package ipc

import (
	"os"
	"path/filepath"
)

// SocketPath returns the anyisland UDS path.
// Override with ANYISLAND_SOCKET, ANYISLAND_IPC_SOCK, or AGENTIC_RUN_DIR.
func SocketPath() string {
	if v := os.Getenv("ANYISLAND_SOCKET"); v != "" {
		return v
	}
	if v := os.Getenv("ANYISLAND_IPC_SOCK"); v != "" {
		return v
	}
	if run := os.Getenv("AGENTIC_RUN_DIR"); run != "" {
		return filepath.Join(run, "anyisland.sock")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".anyisland", "anyisland.sock")
}

// IslandDir returns the anyisland state directory.
func IslandDir() string {
	if v := os.Getenv("ANYISLAND_HOME"); v != "" {
		return v
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".anyisland")
}

// BinDir returns the managed tool install directory.
func BinDir() string {
	if v := os.Getenv("ANYISLAND_BIN_DIR"); v != "" {
		return v
	}
	return filepath.Join(IslandDir(), "bin")
}
