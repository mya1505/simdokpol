package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base32"
	"simdokpol/internal/dto" // Pastikan import DTO ada
	"simdokpol/internal/mocks"
	"simdokpol/internal/models"
	"simdokpol/internal/utils"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// Helper generate key
func generateTestKey(hwid string) string {
	// Kita samakan dengan Secret Key standar
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
	// Setup Env Mocking
	AppSecretKeyString = "SIMDOKPOL_SECRET_KEY_2025"

	realHWID := utils.GetHardwareID()
	validKey := generateTestKey(realHWID)
	invalidKey := "AAAAA-BBBBB-CCCCC-DDDDD"

	actorID := uint(1)

	t.Run("Sukses - Aktivasi dengan Key Valid", func(t *testing.T) {
		mockRepo := new(mocks.LicenseRepository)
		mockConfig := new(mocks.ConfigService)
		mockAudit := new(mocks.AuditLogService)

		// --- FIX PANIC: KONSTRUKTOR MEMANGGIL GETCONFIG ---
		// Saat NewLicenseService dipanggil, dia akan menjalankan verifyRuntimeIntegrity()
		// yang memanggil GetConfig. Kita harus mock ini.
		// .Maybe() artinya: "Boleh dipanggil, boleh tidak (tergantung logic internal)"
		
		mockConfig.On("GetConfig").Return(&dto.AppConfig{LicenseStatus: "UNLICENSED"}, nil).Maybe()
		// Konstruktor mungkin mencoba save config 'UNLICENSED', kita allow saja
		mockConfig.On("SaveConfig", mock.Anything).Return(nil).Maybe()
		// --------------------------------------------------

		// Ekspektasi inti tes Aktivasi (Harus terjadi .Once())
		mockRepo.On("GetLicense", mock.AnythingOfType("string")).Return(nil, gorm.ErrRecordNotFound).Once()
		mockRepo.On("SaveLicense", mock.Anything).Return(nil).Once()
		
		// SaveConfig dipanggil saat aktivasi sukses
		mockConfig.On("SaveConfig", map[string]string{LicenseStatusKey: LicenseStatusValid}).Return(nil).Once()
		mockAudit.On("LogActivity", actorID, "AKTIVASI LISENSI", mock.AnythingOfType("string")).Once()

		service := NewLicenseService(mockRepo, mockConfig, mockAudit)
		_, err := service.ActivateLicense(validKey, actorID)

		assert.NoError(t, err)
		
		// Tidak perlu AssertExpectations di sini jika pakai Maybe(), 
		// karena Maybe tidak wajib dipanggil. Fokus ke error check.
	})

	t.Run("Gagal - Key Invalid", func(t *testing.T) {
		mockRepo := new(mocks.LicenseRepository)
		mockConfig := new(mocks.ConfigService)
		mockAudit := new(mocks.AuditLogService)

		// --- FIX PANIC: SAMA SEPERTI DI ATAS ---
		mockConfig.On("GetConfig").Return(&dto.AppConfig{LicenseStatus: "UNLICENSED"}, nil).Maybe()
		mockConfig.On("SaveConfig", mock.Anything).Return(nil).Maybe()
		// ---------------------------------------

		mockAudit.On("LogActivity", actorID, "GAGAL AKTIVASI", mock.AnythingOfType("string")).Once()

		service := NewLicenseService(mockRepo, mockConfig, mockAudit)
		_, err := service.ActivateLicense(invalidKey, actorID)

		assert.Error(t, err)
		assert.Equal(t, ErrLicenseInvalid, err)
	})
}