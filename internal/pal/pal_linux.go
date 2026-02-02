//go:build linux

package pal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type LinuxPAL struct {
	BasePAL
}

func (p *LinuxPAL) InjectPath() error {
	binDir := p.GetBinDir()
	home, _ := os.UserHomeDir()
	bashrc := filepath.Join(home, ".bashrc")

	content, err := os.ReadFile(bashrc)
	if err != nil {
		return err
	}

	exportCmd := fmt.Sprintf("\nexport PATH=\"" + binDir + ":$PATH\"\n")
	if strings.Contains(string(content), binDir) {
		return nil // Already injected
	}

	f, err := os.OpenFile(bashrc, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString(exportCmd); err != nil {
		return err
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