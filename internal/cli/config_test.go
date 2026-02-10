package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nathfavour/anyisland/internal/pal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigManager(t *testing.T) {
	// Setup a temporary island directory
	tmpDir, err := os.MkdirTemp("", "anyisland-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	mockSys := new(pal.MockSystem)
	mockSys.On("GetIslandDir").Return(tmpDir)

	cm := NewConfigManager(mockSys)

	t.Run("Load default config when file doesn't exist", func(t *testing.T) {
		cfg, err := cm.Load()
		assert.NoError(t, err)
		assert.True(t, cfg.Update.AutoUpdate)
	})

	t.Run("Save and Load config", func(t *testing.T) {
		cfg := &Config{}
		cfg.Update.AutoUpdate = false
		
		err := cm.Save(cfg)
		assert.NoError(t, err)

		// Check if file exists
		assert.FileExists(t, filepath.Join(tmpDir, "config.json"))

		// Load it back
		loaded, err := cm.Load()
		assert.NoError(t, err)
		assert.False(t, loaded.Update.AutoUpdate)
	})

	t.Run("Load invalid json", func(t *testing.T) {
		err := os.WriteFile(filepath.Join(tmpDir, "config.json"), []byte("invalid json"), 0644)
		require.NoError(t, err)

		_, err = cm.Load()
		assert.Error(t, err)
	})
}
