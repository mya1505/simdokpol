package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base32"
	"simdokpol/internal/dto"
	"simdokpol/internal/mocks"
	// "simdokpol/internal/models" // <-- HAPUS ATAU KOMENTARI BARIS INI
	"simdokpol/internal/utils"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// Helper generate key
func generateTestKey(hwid string) string {
	AppSecretKeyString = "SIMDOKPOL_SECRET_KEY_2025"

	h := hmac.New(sha256.New, []byte(AppSecretKeyString))
	h.Write([]byte(hwid))
	hash := h.Sum(nil)
	truncatedHash := hash[:15]
	rawKey := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(truncatedHash)
	
	var formattedKey strings.Builder
	for i, r := range rawKey {
		if i > 0 && i%5 == 0 {
			formattedKey.WriteRune('-')
		}
		formattedKey.WriteRune(r)
	}
	return formattedKey.String()
}

func TestLicenseService_ActivateLicense(t *testing.T) {
	AppSecretKeyString = "SIMDOKPOL_SECRET_KEY_2025"
	realHWID := utils.GetHardwareID()
	validKey := generateTestKey(realHWID)
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