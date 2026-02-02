package pal

import (
	"github.com/zalando/go-keyring"
)

const (
	serviceName = "anyisland"
	accountName = "master-key"
)

type KeyringStore struct{}

func (k *KeyringStore) GetMasterKey() (string, error) {
	return keyring.Get(serviceName, accountName)
}

func (k *KeyringStore) SetMasterKey(key string) error {
	return keyring.Set(serviceName, accountName, key)
}
