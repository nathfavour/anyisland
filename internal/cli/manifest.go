package cli

import (
	"encoding/json"
	"os"

	"github.com/nathfavour/anyisland/internal/agent"
)

type Manifest struct {
	Name    string          `json:"name"`
	Version string          `json:"version"`
	Build   agent.BuildPlan `json:"build"`
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
