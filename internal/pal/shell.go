package pal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func detectShellConfig() (string, error) {
	home, _ := os.UserHomeDir()
	shellPath := os.Getenv("SHELL")
	
	if strings.Contains(shellPath, "zsh") {
		return filepath.Join(home, ".zshrc"), nil
	}
	if strings.Contains(shellPath, "bash") {
		return filepath.Join(home, ".bashrc"), nil
	}
	
	// Fallback to .profile
	return filepath.Join(home, ".profile"), nil
}

func injectPathToConfig(binDir string) error {
	configFile, err := detectShellConfig()
	if err != nil {
		return err
	}

	content, err := os.ReadFile(configFile)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	exportCmd := fmt.Sprintf("\n# Anyisland PATH\nexport PATH=\"%s:$PATH\"\n", binDir)
	if strings.Contains(string(content), binDir) {
		return nil // Already present
	}

	f, err := os.OpenFile(configFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(exportCmd)
	return err
}

func verifyPathInSession(binDir string) bool {
	path := os.Getenv("PATH")
	return strings.Contains(path, binDir)
}
