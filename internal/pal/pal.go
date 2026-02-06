package pal

import (
	"os"
	"path/filepath"
)

// SecretStore defines operations for platform keyring access.
type SecretStore interface {
	GetMasterKey() (string, error)
	SetMasterKey(key string) error
}

// System defines the platform-specific operations required by Anyisland.
type System interface {
	InitFolders() error
	GetIslandDir() string
	GetBinDir() string
	GetIslandBinDir() string
	        GetDataDir() string
	        GetCacheDir() string
	        GetSourceDir() string
	        GetSocketPath() string
	        InjectPath() error
	        EnsurePath() error
	        SecretStore() SecretStore
	}
	
	type BasePAL struct {
	        IslandDir string
	}
	
	func (p *BasePAL) GetIslandDir() string {
	        return p.IslandDir
	}
	
	func (p *BasePAL) GetSocketPath() string {
	        return filepath.Join(p.IslandDir, "anyisland.sock")
	}
	func (p *BasePAL) GetBinDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "bin")
}

func (p *BasePAL) GetIslandBinDir() string {
	return filepath.Join(p.IslandDir, "bin")
}

func (p *BasePAL) GetDataDir() string {
	return filepath.Join(p.IslandDir, "data")
}

func (p *BasePAL) GetCacheDir() string {
	return filepath.Join(p.IslandDir, "cache")
}

func (p *BasePAL) GetSourceDir() string {
	return filepath.Join(p.IslandDir, "source")
}

func (p *BasePAL) InitFolders() error {
	dirs := []string{
		p.GetBinDir(),
		p.GetIslandBinDir(),
		p.GetDataDir(),
		p.GetCacheDir(),
		p.GetSourceDir(),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}

// New returns a platform-specific implementation of System.
func New() (System, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	islandDir := filepath.Join(home, ".anyisland")
	return newSystem(islandDir), nil
}
