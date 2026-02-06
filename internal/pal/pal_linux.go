//go:build linux

package pal

import (
	"fmt"
)

type LinuxPAL struct {
	BasePAL
}

func (p *LinuxPAL) InjectPath() error {
	if err := injectPathToConfig(p.GetBinDir()); err != nil {
		return err
	}
	return injectPathToConfig(p.GetIslandBinDir())
}

func (p *LinuxPAL) EnsurePath() error {
	binDir := p.GetBinDir()
	islandBinDir := p.GetIslandBinDir()
	if !verifyPathInSession(binDir) || !verifyPathInSession(islandBinDir) {
		fmt.Printf("⚠️  Warning: %s or %s is not in your current PATH.\n", binDir, islandBinDir)
		fmt.Println("👉 Please restart your shell or run: source <your-shell-rc-file>")
	}
	return nil
}

func (p *LinuxPAL) SecretStore() SecretStore {
	return &KeyringStore{}
}

func newSystem(islandDir string) System {
	return &LinuxPAL{
		BasePAL: BasePAL{
			IslandDir: islandDir,
		},
	}
}