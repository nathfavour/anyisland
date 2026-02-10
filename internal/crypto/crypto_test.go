package crypto

import (
	"errors"
	"strings"
	"testing"

	"github.com/nathfavour/anyisland/internal/pal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCryptoManager_GetEncryptionKey(t *testing.T) {
	t.Run("Key found in keyring", func(t *testing.T) {
		mockSys := new(pal.MockSystem)
		mockStore := new(pal.MockSecretStore)
		
		mockSys.On("SecretStore").Return(mockStore)
		mockStore.On("GetMasterKey").Return("secret-key", nil)

		cm := &CryptoManager{sys: mockSys}
		key, err := cm.GetEncryptionKey()

		assert.NoError(t, err)
		assert.Equal(t, "secret-key", key)
		mockStore.AssertExpectations(t)
	})

	t.Run("Key not in keyring, prompt user", func(t *testing.T) {
		mockSys := new(pal.MockSystem)
		mockStore := new(pal.MockSecretStore)
		
		mockSys.On("SecretStore").Return(mockStore)
		mockStore.On("GetMasterKey").Return("", errors.New("not found"))
		mockStore.On("SetMasterKey", "user-passphrase").Return(nil)

		input := "user-passphrase
"
		cm := &CryptoManager{
			sys:   mockSys,
			stdin: strings.NewReader(input),
		}
		
		key, err := cm.GetEncryptionKey()

		assert.NoError(t, err)
		assert.Equal(t, "user-passphrase", key)
		mockStore.AssertExpectations(t)
	})
}

func TestCryptoManager_Encrypt(t *testing.T) {
	cm := &CryptoManager{}
	data := "hello world"
	key := "super-secret-key"
	
	encrypted := cm.Encrypt(data, key)
	
	assert.Contains(t, encrypted, "AES256")
	assert.Contains(t, encrypted, data)
	assert.Contains(t, encrypted, "sup...")
}
