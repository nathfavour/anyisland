package pal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInjectPathToConfig(t *testing.T) {
	// Setup temporary home
	tmpHome, err := os.MkdirTemp("", "anyisland-home-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpHome)

	// Mock environment
	oldHome := os.Getenv("HOME")
	oldShell := os.Getenv("SHELL")
	defer func() {
		os.Setenv("HOME", oldHome)
		os.Setenv("SHELL", oldShell)
	}()

	os.Setenv("HOME", tmpHome)
	os.Setenv("SHELL", "/bin/bash")

	bashrc := filepath.Join(tmpHome, ".bashrc")
	binDir := "/path/to/my/bin"

	t.Run("Inject to new file", func(t *testing.T) {
		err := injectPathToConfig(binDir)
		assert.NoError(t, err)

		content, err := os.ReadFile(bashrc)
		assert.NoError(t, err)
		assert.Contains(t, string(content), markerStart)
		assert.Contains(t, string(content), binDir)
		assert.Contains(t, string(content), markerEnd)
	})

	t.Run("Inject duplicate (noop)", func(t *testing.T) {
		err := injectPathToConfig(binDir)
		assert.NoError(t, err)

		content, err := os.ReadFile(bashrc)
		assert.NoError(t, err)
		// Count occurrences of marker
		count := strings.Count(string(content), markerStart)
		assert.Equal(t, 1, count)
	})

	t.Run("Remove from config", func(t *testing.T) {
		err := RemovePathFromConfig()
		assert.NoError(t, err)

		content, err := os.ReadFile(bashrc)
		assert.NoError(t, err)
		assert.NotContains(t, string(content), markerStart)
		assert.NotContains(t, string(content), binDir)
	})
}
