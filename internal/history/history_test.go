package history

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/nathfavour/anyisland/internal/agent"
	"github.com/nathfavour/anyisland/internal/pal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHistoryManager_SyncCommand(t *testing.T) {
	// Setup temporary directory
	tmpDir, err := os.MkdirTemp("", "anyisland-history-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	dataDir := filepath.Join(tmpDir, "data")
	err = os.MkdirAll(dataDir, 0755)
	require.NoError(t, err)

	mockSys := new(pal.MockSystem)
	mockStore := new(pal.MockSecretStore)
	mockAg := new(agent.MockSynthesizer)

	mockSys.On("GetDataDir").Return(dataDir)
	mockSys.On("SecretStore").Return(mockStore)
	mockStore.On("GetMasterKey").Return("test-key", nil)

	mockAg.On("RedactCommand", mock.Anything, "ls -la /home/user").Return("ls -la /home/user", nil)

	hm := NewManager(mockSys, mockAg)

	err = hm.SyncCommand(context.Background(), "ls -la /home/user")
	assert.NoError(t, err)

	// Verify file was created
	historyPath := filepath.Join(dataDir, "history.log")
	assert.FileExists(t, historyPath)

	content, err := os.ReadFile(historyPath)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "AES256")
}
