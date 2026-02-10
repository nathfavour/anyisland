package crypto

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/nathfavour/anyisland/internal/pal"
)

type CryptoManager struct {
	sys   pal.System
	stdin io.Reader
}

func NewManager(sys pal.System) *CryptoManager {
	return &CryptoManager{
		sys:   sys,
		stdin: os.Stdin,
	}
}

func (c *CryptoManager) GetEncryptionKey() (string, error) {
	store := c.sys.SecretStore()
	
	// 1. Try Platform Keyring
	key, err := store.GetMasterKey()
	if err == nil {
		return key, nil
	}

	// 2. Fallback: Ask user for passphrase
	fmt.Println("Platform keyring unavailable or empty.")
	fmt.Print("Enter master passphrase for Anyisland: ")
	reader := bufio.NewReader(c.stdin)
	passphrase, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	passphrase = strings.TrimSpace(passphrase)

	// 3. Persist to keyring for next time (if possible)
	if err := store.SetMasterKey(passphrase); err != nil {
		fmt.Printf("Warning: Could not save to platform keyring: %v\n", err)
	}

	return passphrase, nil
}

func (c *CryptoManager) Encrypt(data string, key string) string {
	// Mock encryption for demonstration
	return fmt.Sprintf("AES256(%s, KEY:%s)", data, key[:3]+"...")
}
