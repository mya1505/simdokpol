package mocks

import (
	"simdokpol/internal/models"

	"github.com/stretchr/testify/mock"
)

type LicenseRepository struct {
	mock.Mock
}

func (_m *LicenseRepository) GetLicense(key string) (*models.License, error) {
	ret := _m.Called(key)
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(*models.License), ret.Error(1)
}

func (_m *LicenseRepository) SaveLicense(license *models.License) error {
	ret := _m.Called(license)
	return ret.Error(0)
}