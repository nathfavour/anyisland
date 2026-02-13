package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadManifest(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "anyisland-manifest-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	manifestPath := filepath.Join(tmpDir, "anyisland.json")

	t.Run("Load valid manifest", func(t *testing.T) {
		content := `{
			"name": "test-tool",
			"version": "1.0.0",
			"build": {
				"steps": ["go build"],
				"bin": "test-tool"
			}
		}`
		err := os.WriteFile(manifestPath, []byte(content), 0644)
		require.NoError(t, err)

		m, err := LoadManifest(manifestPath)
		assert.NoError(t, err)
		assert.Equal(t, "test-tool", m.Name)
		assert.Equal(t, "1.0.0", m.Version)
		assert.Equal(t, "test-tool", m.Build.Bin)
	})

	t.Run("Load manifest with SourceDir", func(t *testing.T) {
		content := `{
			"name": "test-source-dir",
			"version": "1.0.0",
			"source_dir": "~/code/test-tool",
			"build": {
				"steps": ["go build"],
				"bin": "test-tool"
			}
		}`
		err := os.WriteFile(manifestPath, []byte(content), 0644)
		require.NoError(t, err)

		m, err := LoadManifest(manifestPath)
		assert.NoError(t, err)
		assert.Equal(t, "~/code/test-tool", m.SourceDir)
	})

	t.Run("Load manifest with BinName and Aliases", func(t *testing.T) {
		content := `{
			"name": "ripgrep",
			"bin_name": "rg",
			"aliases": ["ripgrep-alias"],
			"version": "13.0.0",
			"build": {
				"steps": ["cargo build --release"],
				"bin": "target/release/rg"
			}
		}`
		err := os.WriteFile(manifestPath, []byte(content), 0644)
		require.NoError(t, err)

		m, err := LoadManifest(manifestPath)
		assert.NoError(t, err)
		assert.Equal(t, "rg", m.BinName)
		assert.Contains(t, m.Aliases, "ripgrep-alias")
	})

	t.Run("Load non-existent file", func(t *testing.T) {
		_, err := LoadManifest("non-existent.json")
		assert.Error(t, err)
	})

	t.Run("Load invalid json", func(t *testing.T) {
		err := os.WriteFile(manifestPath, []byte("invalid json"), 0644)
		require.NoError(t, err)

		_, err = LoadManifest(manifestPath)
		assert.Error(t, err)
	})
}
