package pal

import (
	"github.com/stretchr/testify/mock"
)

type MockSystem struct {
	mock.Mock
}

func (m *MockSystem) InitFolders() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockSystem) GetIslandDir() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockSystem) GetBinDir() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockSystem) GetIslandBinDir() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockSystem) GetDataDir() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockSystem) GetCacheDir() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockSystem) GetSourceDir() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockSystem) GetVisualDir() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockSystem) GetSocketPath() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockSystem) InjectPath() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockSystem) EnsurePath() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockSystem) SecretStore() SecretStore {
	args := m.Called()
	return args.Get(0).(SecretStore)
}

type MockSecretStore struct {
	mock.Mock
}

func (m *MockSecretStore) GetMasterKey() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockSecretStore) SetMasterKey(key string) error {
	args := m.Called(key)
	return args.Error(0)
}
