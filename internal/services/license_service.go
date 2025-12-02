package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base32"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"simdokpol/internal/models"
	"simdokpol/internal/repositories"
	"simdokpol/internal/utils"
	"strings"
	"time"

	"github.com/joho/godotenv" // <-- Pastikan import ini ada
	"gorm.io/gorm"
)

const (
	LicenseStatusKey        = "license_status"
	LicenseStatusValid      = "VALID"
	LicenseStatusUnlicensed = "UNLICENSED"
	EnvLicenseKey           = "LICENSE_KEY"
)

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
	svc.verifyRuntimeIntegrity()
	return svc
}

func (s *licenseService) GetHardwareID() string {
	return utils.GetHardwareID()
}

func (s *licenseService) verifyRuntimeIntegrity() {
	if !s.IsLicensed() {
		_ = s.configService.SaveConfig(map[string]string{LicenseStatusKey: LicenseStatusUnlicensed})
	}
}

func (s *licenseService) IsLicensed() bool {
	// 1. Ambil Key dari Environment
	currentKey := os.Getenv(EnvLicenseKey)

	// --- FIX: FORCE RELOAD .ENV JIKA MEMORY KOSONG ---
	// Kadang env belum termuat sempurna saat service init. Kita paksa baca file.
	if currentKey == "" {
		envPath := filepath.Join(utils.GetAppDataDir(), ".env")
		// Baca manual file .env untuk cari LICENSE_KEY
		if envMap, err := godotenv.Read(envPath); err == nil {
			if keyFromFile, exists := envMap[EnvLicenseKey]; exists && keyFromFile != "" {
				currentKey = keyFromFile
				// Restore ke memory agar request berikutnya cepat
				os.Setenv(EnvLicenseKey, currentKey)
				log.Println("INFO LICENSE: Kunci lisensi dipulihkan dari file .env")
			}
		}
	}
	// -------------------------------------------------

	if currentKey == "" {
		// Jika masih kosong, cek DB. Jika DB Valid tapi Key hilang -> REVOKE (Security)
		status, _ := s.GetLicenseStatus()
		if status == LicenseStatusValid {
			log.Println("SECURITY: Status DB Valid tapi Key hilang. Revoking...")
			_ = s.RevokeLicense()
		}
		return false
	}

	// 2. Verifikasi Kriptografi
	hwid := s.GetHardwareID()
	expectedSignature := generateSignatureRaw(hwid)
	cleanInputKey := strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(currentKey, "-", ""), " ", ""))

	if cleanInputKey == expectedSignature {
		// Valid! Sync DB jika perlu
		status, _ := s.GetLicenseStatus()
		if status != LicenseStatusValid {
			_ = s.configService.SaveConfig(map[string]string{LicenseStatusKey: LicenseStatusValid})
		}
		return true
	}

	// 3. Tidak Valid
	log.Printf("SECURITY: Key invalid terdeteksi. Revoking.")
	_ = s.RevokeLicense()
	return false
}

func (s *licenseService) RevokeLicense() error {
	if err := s.configService.SaveConfig(map[string]string{LicenseStatusKey: LicenseStatusUnlicensed}); err != nil {
		return err
	}
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
	cleanInputKey := strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(inputKey, "-", ""), " ", ""))
	hwid := s.GetHardwareID()
	expectedSignature := generateSignatureRaw(hwid)

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

	existing, _ := s.licenseRepo.GetLicense(cleanInputKey)
	if existing != nil {
		existing.Status = LicenseStatusValid
		existing.ActivatedAt = &now
		existing.ActivatedByID = &actorIDUint
		_ = s.licenseRepo.SaveLicense(existing)
	} else {
		_ = s.licenseRepo.SaveLicense(license)
	}

	if err := s.configService.SaveConfig(map[string]string{LicenseStatusKey: LicenseStatusValid}); err != nil {
		return nil, fmt.Errorf("gagal update config: %w", err)
	}

	// --- FIX: PAKSA TULIS KE FILE .ENV (TANPA SYARAT) ---
	// Kita tidak pakai if os.Getenv != inputKey, tapi langsung tulis biar yakin
	if err := utils.UpdateEnvFile(map[string]string{EnvLicenseKey: inputKey}); err != nil {
		log.Printf("ERROR: Gagal menyimpan lisensi ke file .env: %v", err)
		// Jangan return error fatal, karena di DB sudah aktif. Cuma warning log.
	} else {
		log.Println("SUCCESS: Lisensi tersimpan permanen di file .env")
	}
	os.Setenv(EnvLicenseKey, inputKey)
	// ---------------------------------------------------

	if actorID != 0 {
		s.auditService.LogActivity(actorID, "AKTIVASI LISENSI", "Lisensi PRO berhasil diaktifkan.")
	}
	return license, nil
}

func generateSignatureRaw(data string) string {
	h := hmac.New(sha256.New, []byte(AppSecretKeyString))
	h.Write([]byte(data))
	hash := h.Sum(nil)
	truncatedHash := hash[:15]
	encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(truncatedHash)
	return encoded
}