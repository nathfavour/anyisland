//go:build windows

package pal

import (
	"os"
	"os/exec"
	"strings"
)

type WindowsPAL struct {
	BasePAL
}

func (p *WindowsPAL) InjectPath() error {
	binDir := p.GetBinDir()
	islandBinDir := p.GetIslandBinDir()
	path := os.Getenv("PATH")
	
	newPath := path
	changed := false
	if !strings.Contains(path, binDir) {
		newPath = binDir + ";" + newPath
		changed = true
	}
	if !strings.Contains(path, islandBinDir) {
		newPath = islandBinDir + ";" + newPath
		changed = true
	}

	if !changed {
		return nil
	}

	// Use setx to permanently update PATH
	cmd := exec.Command("setx", "PATH", newPath)
	return cmd.Run()
}

func (p *WindowsPAL) EnsurePath() error {
	binDir := p.GetBinDir()
	islandBinDir := p.GetIslandBinDir()
	if !verifyPathInSession(binDir) || !verifyPathInSession(islandBinDir) {
		fmt.Printf("⚠️  Warning: %s or %s is not in your current PATH.\n", binDir, islandBinDir)
		fmt.Println("👉 You may need to restart your terminal or computer for environment changes to take effect.")
	}
	return nil
}

func (p *WindowsPAL) SecretStore() SecretStore {
	return &KeyringStore{}
}

func newSystem(islandDir string) System {
	return &WindowsPAL{
		BasePAL: BasePAL{
			IslandDir: islandDir,
		},
	}
}