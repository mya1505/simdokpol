package services

import (
	"os"
	"path/filepath"
	"simdokpol/internal/dto"
	"simdokpol/internal/mocks"
	"simdokpol/internal/utils"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// Helper generate key
func loadTestPrivateKey(t *testing.T) []byte {
	t.Helper()

	paths := []string{
		"private.pem",
		filepath.Join("..", "..", "private.pem"),
	}

	for _, path := range paths {
		if pemBytes, err := os.ReadFile(path); err == nil {
			return pemBytes
		}
	}
	t.Fatal("private.pem tidak ditemukan untuk test aktivasi")
	return nil
}

func generateTestKey(t *testing.T, hwid string) string {
	t.Helper()

	pemBytes := loadTestPrivateKey(t)
	privateKey, err := utils.ParsePrivateKeyPEM(pemBytes)
	if err != nil {
		t.Fatalf("gagal parse private key test: %v", err)
	}

	key, err := utils.SignActivationKey(hwid, privateKey)
	if err != nil {
		t.Fatalf("gagal sign activation key test: %v", err)
	}
	return key
}

func TestLicenseService_ActivateLicense(t *testing.T) {
	realHWID := utils.GetHardwareID()
	validKey := generateTestKey(t, realHWID)
	invalidKey := "AAAAA-BBBBB-CCCCC-DDDDD"
	actorID := uint(1)

	t.Run("Sukses - Aktivasi dengan Key Valid", func(t *testing.T) {
		mockRepo := new(mocks.LicenseRepository)
		mockConfig := new(mocks.ConfigService)
		mockAudit := new(mocks.AuditLogService)

		mockConfig.On("GetConfig").Return(&dto.AppConfig{LicenseStatus: "UNLICENSED"}, nil).Maybe()
		mockConfig.On("SaveConfig", mock.Anything).Return(nil).Maybe()

		mockRepo.On("GetLicense", mock.AnythingOfType("string")).Return(nil, gorm.ErrRecordNotFound).Once()
		mockRepo.On("SaveLicense", mock.Anything).Return(nil).Once()

		mockConfig.On("SaveConfig", map[string]string{"license_status": "VALID"}).Return(nil).Once()
		mockAudit.On("LogActivity", actorID, "AKTIVASI LISENSI", mock.AnythingOfType("string")).Once()

		service := NewLicenseService(mockRepo, mockConfig, mockAudit)
		_, err := service.ActivateLicense(validKey, actorID)

		assert.NoError(t, err)
	})

	t.Run("Gagal - Key Invalid", func(t *testing.T) {
		mockRepo := new(mocks.LicenseRepository)
		mockConfig := new(mocks.ConfigService)
		mockAudit := new(mocks.AuditLogService)

		mockConfig.On("GetConfig").Return(&dto.AppConfig{LicenseStatus: "UNLICENSED"}, nil).Maybe()
		mockConfig.On("SaveConfig", mock.Anything).Return(nil).Maybe()

		mockAudit.On("LogActivity", actorID, "GAGAL AKTIVASI", mock.AnythingOfType("string")).Once()

		service := NewLicenseService(mockRepo, mockConfig, mockAudit)
		_, err := service.ActivateLicense(invalidKey, actorID)

		assert.Error(t, err)
		assert.Equal(t, ErrLicenseInvalid, err)
	})
}
