//go:build windows

package pal

type WindowsPAL struct {
	BasePAL
}

func newSystem(islandDir string) System {
	return &WindowsPAL{
		BasePAL: BasePAL{
			IslandDir: islandDir,
		},
	}
}
