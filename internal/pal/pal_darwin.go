//go:build darwin

package pal

import (
	"fmt"
)

type DarwinPAL struct {
	BasePAL
}

func (p *DarwinPAL) InjectPath() error {
	if err := injectPathToConfig(p.GetBinDir()); err != nil {
		return err
	}
	return injectPathToConfig(p.GetIslandBinDir())
}

func (p *DarwinPAL) EnsurePath() error {
	binDir := p.GetBinDir()
	islandBinDir := p.GetIslandBinDir()
	if !verifyPathInSession(binDir) || !verifyPathInSession(islandBinDir) {
		fmt.Printf("⚠️  Warning: %s or %s is not in your current PATH.\n", binDir, islandBinDir)
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
