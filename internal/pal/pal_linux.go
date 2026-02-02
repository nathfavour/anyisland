//go:build linux

package pal

type LinuxPAL struct {
	BasePAL
}

func newSystem(islandDir string) System {
	return &LinuxPAL{
		BasePAL: BasePAL{
			IslandDir: islandDir,
		},
	}
}
