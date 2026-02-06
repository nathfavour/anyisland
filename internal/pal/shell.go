package pal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	markerStart = "# >>> anyisland initialize >>>"
	markerEnd   = "# <<< anyisland initialize <<<"
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

	content := ""
	data, err := os.ReadFile(configFile)
	if err == nil {
		content = string(data)
	} else if !os.IsNotExist(err) {
		return err
	}

	// Check if our markers already exist
	if strings.Contains(content, markerStart) {
		// If the binDir is already in the file, we assume it's fine.
		// A more robust implementation would update the block if needed.
		if strings.Contains(content, binDir) {
			return nil
		}
		// Remove old block to re-inject
		if err := RemovePathFromConfig(); err != nil {
			return err
		}
		// Reload content
		data, _ = os.ReadFile(configFile)
		content = string(data)
	}

	exportCmd := fmt.Sprintf("\n%s\nexport PATH=\"$PATH:%s\"\n%s\n", markerStart, binDir, markerEnd)

	f, err := os.OpenFile(configFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(exportCmd)
	return err
}

func RemovePathFromConfig() error {
	configFile, err := detectShellConfig()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	content := string(data)
	if !strings.Contains(content, markerStart) {
		return nil
	}

	lines := strings.Split(content, "\n")
	var newLines []string
	skipping := false

	for _, line := range lines {
		if strings.Contains(line, markerStart) {
			skipping = true
			continue
		}
		if strings.Contains(line, markerEnd) {
			skipping = false
			continue
		}
		if !skipping {
			newLines = append(newLines, line)
		}
	}

	return os.WriteFile(configFile, []byte(strings.Join(newLines, "\n")), 0644)
}

func verifyPathInSession(binDir string) bool {
	path := os.Getenv("PATH")
	return strings.Contains(path, binDir)
}