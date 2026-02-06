package cli

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/nathfavour/anyisland/internal/agent"
)

func InitProject(dir string) error {
	absDir, _ := filepath.Abs(dir)
	name := filepath.Base(absDir)
	path := filepath.Join(absDir, "anyisland.json")
	m := Manifest{
		Name:        name,
		Version:     "0.1.0",
		Description: "A new tool managed by Anyisland.",
		InstallDir:  "", // Optional: override default installation path (~/.anyisland/bin)
		Build: agent.BuildPlan{
			Steps: []string{"go build -o " + name},
			Bin:   name,
		},
		Runtime: &RuntimeConfig{
			Dependencies: []string{},
			Daemon:       false,
			Pulse:        false,
		},
	}

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
