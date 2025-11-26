package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base32"
	"simdokpol/internal/mocks"
	"simdokpol/internal/utils"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// Helper untuk membuat key valid di dalam test
func generateTestKey(hwid string) string {
	// Set Key Global dulu untuk keperluan testing
	AppSecretKeyString = "TEST-SECRET-KEY-123" 

	h := hmac.New(sha256.New, []byte(AppSecretKeyString)) // <-- Gunakan variabel global
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
	realHWID := utils.GetHardwareID()
	validKey := generateTestKey(realHWID)
	invalidKey := "AAAAA-BBBBB-CCCCC-DDDDD"

	actorID := uint(1)

	t.Run("Sukses - Aktivasi dengan Key Valid", func(t *testing.T) {
		mockRepo := new(mocks.LicenseRepository)
		mockConfig := new(mocks.ConfigService)
		mockAudit := new(mocks.AuditLogService)

		mockRepo.On("GetLicense", mock.AnythingOfType("string")).Return(nil, gorm.ErrRecordNotFound).Once()
		
		mockRepo.On("SaveLicense", mock.Anything).Return(nil).Once()
		
		mockConfig.On("SaveConfig", map[string]string{LicenseStatusKey: LicenseStatusValid}).Return(nil).Once()
		mockAudit.On("LogActivity", actorID, "AKTIVASI LISENSI", mock.AnythingOfType("string")).Once()

		service := NewLicenseService(mockRepo, mockConfig, mockAudit)
		_, err := service.ActivateLicense(validKey, actorID)

		assert.NoError(t, err)
		
		mockRepo.AssertExpectations(t)
		mockConfig.AssertExpectations(t)
		mockAudit.AssertExpectations(t)
	})

	t.Run("Gagal - Key Invalid", func(t *testing.T) {
		mockRepo := new(mocks.LicenseRepository)
		mockConfig := new(mocks.ConfigService)
		mockAudit := new(mocks.AuditLogService)

		mockAudit.On("LogActivity", actorID, "GAGAL AKTIVASI", mock.AnythingOfType("string")).Once()

		service := NewLicenseService(mockRepo, mockConfig, mockAudit)
		_, err := service.ActivateLicense(invalidKey, actorID)

		assert.Error(t, err)
		assert.Equal(t, ErrLicenseInvalid, err)
		
		// Verifikasi
		mockRepo.AssertExpectations(t)
		mockAudit.AssertExpectations(t)
	})
}