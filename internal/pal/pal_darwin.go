//go:build darwin

package pal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type DarwinPAL struct {
	BasePAL
}

func (p *DarwinPAL) InjectPath() error {
	home, _ := os.UserHomeDir()
	zshrc := filepath.Join(home, ".zshrc")

	content, err := os.ReadFile(zshrc)
	if err != nil {
		return err
	}

	exportCmd := fmt.Sprintf("\nexport PATH=\"$HOME/.local/bin:$PATH\"\n")
	if strings.Contains(string(content), ".local/bin") {
		return nil // Already injected
	}

	f, err := os.OpenFile(zshrc, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString(exportCmd); err != nil {
		return err
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