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
	path := os.Getenv("PATH")
	if strings.Contains(path, binDir) {
		return nil
	}

	// Use setx to permanently update PATH
	cmd := exec.Command("setx", "PATH", path+";"+binDir)
	return cmd.Run()
}

func (p *WindowsPAL) EnsurePath() error {
	binDir := p.GetBinDir()
	if !verifyPathInSession(binDir) {
		fmt.Printf("⚠️  Warning: %s is not in your current PATH.\n", binDir)
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