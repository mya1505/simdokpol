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

const LicenseStatusKey = "license_status"
const LicenseStatusValid = "VALID"
const LicenseStatusUnlicensed = "UNLICENSED"
const EnvLicenseKey = "LICENSE_KEY"

// FIX B-01: Ubah const menjadi var agar bisa diinjeksi via ldflags saat build
// Default value kosong untuk keamanan. Jangan pernah commit key asli di sini!
var AppSecretKeyString = ""

var ErrLicenseInvalid = errors.New("kunci lisensi tidak valid untuk mesin ini")
var ErrLicenseBanned = errors.New("kunci lisensi ini telah diblokir")

type LicenseService interface {
	ActivateLicense(key string, actorID uint) (*models.License, error)
	GetLicenseStatus() (string, error)
	IsLicensed() bool
	GetHardwareID() string
	AutoActivateFromEnv()
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
	svc.AutoActivateFromEnv()
	return svc
}

func (s *licenseService) GetHardwareID() string {
	return utils.GetHardwareID()
}

func (s *licenseService) IsLicensed() bool {
	status, err := s.GetLicenseStatus()
	if err != nil {
		return false
	}
	return status == LicenseStatusValid
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

func (s *licenseService) AutoActivateFromEnv() {
	envKey := os.Getenv(EnvLicenseKey)
	if envKey == "" {
		return
	}

	if s.IsLicensed() {
		return
	}

	log.Println("INFO: Mencoba auto-aktivasi lisensi dari .env...")
	_, err := s.ActivateLicense(envKey, 0)
	if err != nil {
		log.Printf("WARNING: Auto-aktivasi lisensi gagal: %v", err)
	} else {
		log.Println("SUCCESS: Lisensi berhasil dipulihkan dari .env!")
	}
}

func (s *licenseService) ActivateLicense(inputKey string, actorID uint) (*models.License, error) {
	// FIX B-01: Validasi bahwa AppSecretKeyString telah di-set saat build
	if AppSecretKeyString == "" {
		log.Println("CRITICAL: AppSecretKeyString belum diset! Gunakan -ldflags saat build.")
		return nil, errors.New("konfigurasi keamanan server belum diinisialisasi")
	}

	rawInput := inputKey
	cleanKey := strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(inputKey, "-", ""), " ", ""))

	hwid := s.GetHardwareID()
	expectedSignature := generateSignature(hwid)

	if cleanKey != expectedSignature {
		if actorID != 0 {
			s.auditService.LogActivity(actorID, "GAGAL AKTIVASI", fmt.Sprintf("Key salah. Input: %s, HWID: %s", rawInput, hwid))
		}
		return nil, ErrLicenseInvalid
	}

	now := time.Now()
	actorIDUint := actorID

	license := &models.License{
		Key:           cleanKey,
		Status:        LicenseStatusValid,
		ActivatedAt:   &now,
		ActivatedByID: &actorIDUint,
		Notes:         fmt.Sprintf("Aktivasi sukses. HWID: %s", hwid),
	}

	existing, _ := s.licenseRepo.GetLicense(cleanKey)
	if existing != nil {
		existing.Status = LicenseStatusValid
		existing.ActivatedAt = &now
		existing.ActivatedByID = &actorIDUint
		if err := s.licenseRepo.SaveLicense(existing); err != nil {
			return nil, err
		}
	} else {
		if err := s.licenseRepo.SaveLicense(license); err != nil {
			return nil, fmt.Errorf("gagal menyimpan lisensi: %w", err)
		}
	}

	if err := s.configService.SaveConfig(map[string]string{LicenseStatusKey: LicenseStatusValid}); err != nil {
		return nil, fmt.Errorf("gagal update status konfigurasi: %w", err)
	}

	if os.Getenv(EnvLicenseKey) != cleanKey {
		if err := utils.UpdateEnvFile(map[string]string{EnvLicenseKey: cleanKey}); err != nil {
			log.Printf("WARNING: Gagal menyimpan license key ke .env: %v", err)
		}
	}

	if actorID != 0 {
		s.auditService.LogActivity(actorID, "AKTIVASI LISENSI", "Lisensi PRO berhasil diaktifkan.")
	}
	return license, nil
}

func generateSignature(data string) string {
	// FIX B-01: Gunakan variable global
	h := hmac.New(sha256.New, []byte(AppSecretKeyString))
	h.Write([]byte(data))
	hash := h.Sum(nil)
	truncatedHash := hash[:15]
	encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(truncatedHash)
	return encoded
}
