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

func newSystem(islandDir string) System {
	return &WindowsPAL{
		BasePAL: BasePAL{
			IslandDir: islandDir,
		},
	}
}