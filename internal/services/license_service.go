package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base32"
	"errors"
	"fmt"
	"log"
	"os"
	"simdokpol/internal/models"
	"simdokpol/internal/repositories"
	"simdokpol/internal/utils"
	"strings"
	"time"

	"gorm.io/gorm"
)

const (
	LicenseStatusKey        = "license_status"
	LicenseStatusValid      = "VALID"
	LicenseStatusUnlicensed = "UNLICENSED"
	EnvLicenseKey           = "LICENSE_KEY"
)

// Kunci Rahasia Aplikasi (Di Production ini harus di-inject via ldflags saat build)
// Makefile akan menimpa variabel ini.
var AppSecretKeyString = "JANGAN_PAKAI_DEFAULT_KEY_INI_BAHAYA"

var (
	ErrLicenseInvalid = errors.New("kunci lisensi tidak valid untuk mesin ini")
	ErrLicenseBanned  = errors.New("kunci lisensi ini telah diblokir")
)

type LicenseService interface {
	ActivateLicense(key string, actorID uint) (*models.License, error)
	GetLicenseStatus() (string, error)
	IsLicensed() bool
	GetHardwareID() string
	RevokeLicense() error
}

type licenseService struct {
	licenseRepo   repositories.LicenseRepository
	configService ConfigService
	auditService  AuditLogService
}

func NewLicenseService(licenseRepo repositories.LicenseRepository, configService ConfigService, auditService AuditLogService) LicenseService {
	svc := &licenseService{
		licenseRepo:   licenseRepo,
		configService: configService,
		auditService:  auditService,
	}
	// Cek integritas saat startup
	svc.verifyRuntimeIntegrity()
	return svc
}

func (s *licenseService) GetHardwareID() string {
	return utils.GetHardwareID()
}

func (s *licenseService) verifyRuntimeIntegrity() {
	// Jika menurut logika IsLicensed() tidak valid, paksa status DB jadi UNLICENSED
	if !s.IsLicensed() {
		_ = s.configService.SaveConfig(map[string]string{LicenseStatusKey: LicenseStatusUnlicensed})
	}
}

// IsLicensed: Gatekeeper utama dengan Zero Trust
func (s *licenseService) IsLicensed() bool {
	// 1. Ambil Key dari Environment (Memori/File)
	currentKey := os.Getenv(EnvLicenseKey)

	// Jika tidak ada key di environment, anggap tidak berlisensi
	if currentKey == "" {
		// Jika DB masih bilang VALID, kita revoke karena suspicious
		status, _ := s.GetLicenseStatus()
		if status == LicenseStatusValid {
			_ = s.RevokeLicense()
		}
		return false
	}

	// 2. Verifikasi Kriptografi
	hwid := s.GetHardwareID()
	// Generate hash yang diharapkan (tanpa format dash)
	expectedSignature := generateSignatureRaw(hwid)

	// Bersihkan input key (buang dash, spasi, uppercase)
	cleanInputKey := strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(currentKey, "-", ""), " ", ""))

	// Bandingkan
	if cleanInputKey == expectedSignature {
		// Valid! Pastikan DB sync
		status, _ := s.GetLicenseStatus()
		if status != LicenseStatusValid {
			_ = s.configService.SaveConfig(map[string]string{LicenseStatusKey: LicenseStatusValid})
		}
		return true
	}

	// 3. Tidak Valid (Key ada tapi salah)
	log.Printf("SECURITY: Key invalid terdeteksi. Revoking status Pro.")
	_ = s.RevokeLicense()
	return false
}

func (s *licenseService) RevokeLicense() error {
	// Set DB ke UNLICENSED
	if err := s.configService.SaveConfig(map[string]string{LicenseStatusKey: LicenseStatusUnlicensed}); err != nil {
		return err
	}
	// Bersihkan Env var
	os.Setenv(EnvLicenseKey, "")
	_ = utils.UpdateEnvFile(map[string]string{EnvLicenseKey: ""})
	return nil
}

func (s *licenseService) GetLicenseStatus() (string, error) {
	config, err := s.configService.GetConfig()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return LicenseStatusUnlicensed, nil
		}
		return "", err
	}
	if config.LicenseStatus == "" {
		return LicenseStatusUnlicensed, nil
	}
	return config.LicenseStatus, nil
}

func (s *licenseService) ActivateLicense(inputKey string, actorID uint) (*models.License, error) {
	// Bersihkan input
	cleanInputKey := strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(inputKey, "-", ""), " ", ""))

	hwid := s.GetHardwareID()
	expectedSignature := generateSignatureRaw(hwid)

	// Validasi
	if cleanInputKey != expectedSignature {
		if actorID != 0 {
			s.auditService.LogActivity(actorID, "GAGAL AKTIVASI", fmt.Sprintf("Key salah. HWID: %s", hwid))
		}
		return nil, ErrLicenseInvalid
	}

	now := time.Now()
	actorIDUint := actorID

	license := &models.License{
		Key:           cleanInputKey,
		Status:        LicenseStatusValid,
		ActivatedAt:   &now,
		ActivatedByID: &actorIDUint,
		Notes:         fmt.Sprintf("Aktivasi sukses. HWID: %s", hwid),
	}

	// Simpan/Update tabel license
	existing, _ := s.licenseRepo.GetLicense(cleanInputKey)
	if existing != nil {
		existing.Status = LicenseStatusValid
		existing.ActivatedAt = &now
		existing.ActivatedByID = &actorIDUint
		_ = s.licenseRepo.SaveLicense(existing)
	} else {
		_ = s.licenseRepo.SaveLicense(license)
	}

	// Update Config DB
	if err := s.configService.SaveConfig(map[string]string{LicenseStatusKey: LicenseStatusValid}); err != nil {
		return nil, fmt.Errorf("gagal update config: %w", err)
	}

	// Update .env & Memori agar persist
	// Simpan key ASLI input user (yang ada dash-nya lebih mudah dibaca di file)
	_ = utils.UpdateEnvFile(map[string]string{EnvLicenseKey: inputKey})
	os.Setenv(EnvLicenseKey, inputKey)

	if actorID != 0 {
		s.auditService.LogActivity(actorID, "AKTIVASI LISENSI", "Lisensi PRO berhasil diaktifkan.")
	}

	log.Printf("SUCCESS: License activated for HWID: %s", hwid)
	return license, nil
}

// generateSignatureRaw: Membuat hash MENTAH (tanpa dash) untuk verifikasi internal
func generateSignatureRaw(data string) string {
	h := hmac.New(sha256.New, []byte(AppSecretKeyString))
	h.Write([]byte(data))
	hash := h.Sum(nil)

	// Ambil 15 byte pertama
	truncatedHash := hash[:15]
	// Encode Base32 (Tanpa Padding)
	encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(truncatedHash)

	return encoded
}