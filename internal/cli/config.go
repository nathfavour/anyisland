package cli

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/nathfavour/anyisland/internal/pal"
)

type Config struct {
	Update struct {
		AutoUpdate bool `json:"auto_update"`
	} `json:"update"`
	Install struct {
		Preference    string `json:"preference"`     // "binary" or "source"
		DefaultBranch string `json:"default_branch"` // e.g. "main", "master", or empty for repo default
	} `json:"install"`
}

type ConfigManager struct {
	sys pal.System
}

func NewConfigManager(sys pal.System) *ConfigManager {
	return &ConfigManager{sys: sys}
}

func (cm *ConfigManager) GetConfigPath() string {
	return filepath.Join(cm.sys.GetIslandDir(), "config.json")
}

func (cm *ConfigManager) Load() (*Config, error) {
	path := cm.GetConfigPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Return default config
		cfg := &Config{}
		cfg.Update.AutoUpdate = true
		cfg.Install.Preference = "binary" // Default to binary for speed
		cfg.Install.DefaultBranch = ""     // Default to repo's default branch
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Apply defaults for missing fields
	if cfg.Install.Preference == "" {
		cfg.Install.Preference = "binary"
	}

	return &cfg, nil
}

func (cm *ConfigManager) Save(cfg *Config) error {
	path := cm.GetConfigPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
