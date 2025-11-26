package mocks

import (
	"simdokpol/internal/models"

	"github.com/stretchr/testify/mock"
)

type LicenseService struct {
	mock.Mock
}

func (_m *LicenseService) ActivateLicense(key string, actorID uint) (*models.License, error) {
	ret := _m.Called(key, actorID)
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(*models.License), ret.Error(1)
}

func (_m *LicenseService) GetLicenseStatus() (string, error) {
	ret := _m.Called()
	return ret.String(0), ret.Error(1)
}

func (_m *LicenseService) IsLicensed() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *LicenseService) GetHardwareID() string {
	ret := _m.Called()
	return ret.String(0)
}

// --- METODE BARU YANG HILANG ---
func (_m *LicenseService) AutoActivateFromEnv() {
	_m.Called()
}
// -------------------------------