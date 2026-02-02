//go:build linux

package pal

import (
	"fmt"
)

type LinuxPAL struct {
	BasePAL
}

func (p *LinuxPAL) InjectPath() error {
	return injectPathToConfig(p.GetBinDir())
}

func (p *LinuxPAL) EnsurePath() error {
	binDir := p.GetBinDir()
	if !verifyPathInSession(binDir) {
		fmt.Printf("⚠️  Warning: %s is not in your current PATH.\n", binDir)
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