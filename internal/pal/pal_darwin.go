//go:build darwin

package pal

type DarwinPAL struct {
	BasePAL
}

func newSystem(islandDir string) System {
	return &DarwinPAL{
		BasePAL: BasePAL{
			IslandDir: islandDir,
		},
	}
}
