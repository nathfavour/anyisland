package cli

import (
	"encoding/json"
	"os"

	"github.com/nathfavour/anyisland/internal/agent"
)

type RuntimeConfig struct {
	Daemon         bool   `json:"daemon,omitempty"`
	Pulse          bool   `json:"pulse,omitempty"`
	ManagedUpdates *bool  `json:"managed_updates,omitempty"` // If false, Anyisland won't auto-update this tool
	UpdateCommand  string `json:"update_command,omitempty"`  // Custom command to update the tool
}

type ReleaseInfo struct {
	TagName     string `json:"tag_name,omitempty"`
	AssetURL    string `json:"asset_url,omitempty"`
	AssetName   string `json:"asset_name,omitempty"`
	IsBinary    bool   `json:"is_binary,omitempty"`
	PublishedAt string `json:"published_at,omitempty"`
}

type FeatureConfig struct {
	TrackSubmodules  bool `json:"track_submodules,omitempty"`
	RecursiveInstall bool `json:"recursive_install,omitempty"`
}

type Manifest struct {
	Name         string          `json:"name"`
	BinName      string          `json:"bin_name,omitempty"` // The actual binary name to be installed in PATH
	Aliases      []string        `json:"aliases,omitempty"`  // Alternative names/symlinks for the tool
	Version      string          `json:"version"`
	Description  string          `json:"description,omitempty"`
	Repository   string          `json:"repository,omitempty"`
	Branch       string          `json:"branch,omitempty"`
	Dependencies []string        `json:"dependencies,omitempty"`
	SourceDir    string          `json:"source_dir,omitempty"`  // Custom path to store or find source code
	InstallDir   string          `json:"install_dir,omitempty"` // Root-level override for installation path
	Build        agent.BuildPlan `json:"build"`
	Runtime      *RuntimeConfig  `json:"runtime,omitempty"`
	Features     *FeatureConfig  `json:"features,omitempty"`
	Release      *ReleaseInfo    `json:"release,omitempty"` // Metadata if installed from a release
}

func LoadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}
