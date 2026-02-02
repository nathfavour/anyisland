//go:build darwin

package pal

import (
	"fmt"
)

type DarwinPAL struct {
	BasePAL
}

func (p *DarwinPAL) InjectPath() error {
	return injectPathToConfig(p.GetBinDir())
}

func (p *DarwinPAL) EnsurePath() error {
	binDir := p.GetBinDir()
	if !verifyPathInSession(binDir) {
		fmt.Printf("⚠️  Warning: %s is not in your current PATH.\n", binDir)
		fmt.Println("👉 Please restart your shell or run: source <your-shell-rc-file>")
	}
	return nil
}

func (p *DarwinPAL) SecretStore() SecretStore {
	return &KeyringStore{}
}

func newSystem(islandDir string) System {
	return &DarwinPAL{
		BasePAL: BasePAL{
			IslandDir: islandDir,
		},
	}
}