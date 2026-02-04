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
		Name:    name,
		Version: "0.1.0",
		Build: agent.BuildPlan{
			Steps: []string{"go build -o " + name},
			Bin:   name,
		},
	}

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
