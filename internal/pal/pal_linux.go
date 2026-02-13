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

	binInPath := verifyPathInSession(binDir)
	islandBinInPath := verifyPathInSession(islandBinDir)

	if !binInPath || !islandBinInPath {
		var missing []string
		if !binInPath {
			missing = append(missing, binDir)
		}
		if !islandBinInPath {
			missing = append(missing, islandBinDir)
		}

		fmt.Printf("⚠️  Warning: The following directories are not in your PATH:\n")
		for _, m := range missing {
			fmt.Printf("   - %s\n", m)
		}
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
