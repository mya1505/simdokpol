package mocks

import (
	"simdokpol/internal/models"

	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// ConfigRepository adalah mock untuk repositories.ConfigRepository
type ConfigRepository struct {
	mock.Mock
}

func (m *ConfigRepository) Get(key string) (*models.Configuration, error) {
	args := m.Called(key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Configuration), args.Error(1)
}

func (m *ConfigRepository) GetForUpdate(tx *gorm.DB, key string) (*models.Configuration, error) {
	args := m.Called(tx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Configuration), args.Error(1)
}

func (m *ConfigRepository) GetAll() (map[string]string, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]string), args.Error(1)
}

func (m *ConfigRepository) Set(key string, value string) error {
	args := m.Called(key, value)
	return args.Error(0)
}

func (m *ConfigRepository) SetMultiple(tx *gorm.DB, configs map[string]string) error {
	args := m.Called(tx, configs)
	return args.Error(0)
}