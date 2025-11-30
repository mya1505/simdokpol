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

// Default Key (Hanya untuk Development, Production di-override Makefile)
var AppSecretKeyString = "DEFAULT_DEV_KEY_JANGAN_DIPAKAI"

var ErrLicenseInvalid = errors.New("kunci lisensi tidak valid untuk mesin ini")
var ErrLicenseBanned = errors.New("kunci lisensi ini telah diblokir")

type LicenseService interface {
	ActivateLicense(key string, actorID uint) (*models.License, error)
	GetLicenseStatus() (string, error)
	IsLicensed() bool
	GetHardwareID() string
	AutoActivateFromEnv()
	RevokeLicense() error // <-- Method Baru
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

// --- FIX KRITIS: VALIDASI ULANG SAAT RUNTIME ---
func (s *licenseService) IsLicensed() bool {
	// 1. Cek apa kata Config/DB (Cepat)
	status, err := s.GetLicenseStatus()
	if err != nil || status != LicenseStatusValid {
		return false
	}

	// 2. SECURITY CHECK (Audit Integritas)
	// Jangan percaya status DB buta-buta. Kita cek kuncinya ada nggak di Env/DB?
	currentKey := os.Getenv(EnvLicenseKey)
	
	// Jika Env kosong, coba cari license aktif terakhir di DB
	if currentKey == "" {
		// Logic ini opsional, tapi bagus buat fallback
		// Namun biar aman, kita anggap kalau Env kosong = suspicious
	}

	if currentKey == "" {
		// Status VALID tapi Key gak ada? Aneh. Revoke!
		// log.Println("SECURITY: Status VALID tapi Key kosong. Revoking...")
		// s.RevokeLicense() 
		// (Opsional: uncomment revoke jika ingin strict, tapi return false cukup)
		return false
	}

	// 3. Re-Verify HMAC (Paling Penting)
	// Pastikan key yang tersimpan emang cocok buat HWID mesin ini.
	// Ini mencegah orang copy-paste file .db atau .env dari komputer lain yang udah Pro.
	hwid := s.GetHardwareID()
	expectedSignature := generateSignature(hwid)
	cleanKey := strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(currentKey, "-", ""), " ", ""))

	if cleanKey != expectedSignature {
		log.Printf("SECURITY ALERT: Lisensi corrupt/palsu terdeteksi! HWID: %s", hwid)
		s.RevokeLicense() // Otomatis matikan Pro
		return false
	}

	return true
}

// Helper untuk membatalkan lisensi (Fallback ke Free)
func (s *licenseService) RevokeLicense() error {
	// Update DB Config
	if err := s.configService.SaveConfig(map[string]string{LicenseStatusKey: LicenseStatusUnlicensed}); err != nil {
		return err
	}
	// Hapus dari .Env (biar gak reload lagi pas restart)
	utils.UpdateEnvFile(map[string]string{EnvLicenseKey: ""})
	os.Setenv(EnvLicenseKey, "") // Clear memory env
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

func (s *licenseService) AutoActivateFromEnv() {
	envKey := os.Getenv(EnvLicenseKey)
	if envKey == "" {
		return
	}

	// Jangan cek IsLicensed() di sini biar gak infinite loop, 
	// langsung validasi key-nya aja.
	
	cleanKey := strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(envKey, "-", ""), " ", ""))
	hwid := s.GetHardwareID()
	expected := generateSignature(hwid)

	if cleanKey == expected {
		// Jika cocok, pastikan DB sync
		s.configService.SaveConfig(map[string]string{LicenseStatusKey: LicenseStatusValid})
	} else {
		// Jika ada key di .env tapi salah (misal ganti hardware), hapus.
		log.Println("INFO: Key di .env tidak valid untuk mesin ini. Menghapus...")
		s.RevokeLicense()
	}
}

func (s *licenseService) ActivateLicense(inputKey string, actorID uint) (*models.License, error) {
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

	// Simpan ke .env agar persist saat restart
	if os.Getenv(EnvLicenseKey) != rawInput {
		if err := utils.UpdateEnvFile(map[string]string{EnvLicenseKey: rawInput}); err != nil {
			log.Printf("WARNING: Gagal menyimpan license key ke .env: %v", err)
		}
		// Set juga di memori untuk sesi ini
		os.Setenv(EnvLicenseKey, rawInput)
	}

	if actorID != 0 {
		s.auditService.LogActivity(actorID, "AKTIVASI LISENSI", "Lisensi PRO berhasil diaktifkan.")
	}
	return license, nil
}

func generateSignature(data string) string {
	h := hmac.New(sha256.New, []byte(AppSecretKeyString))
	h.Write([]byte(data))
	hash := h.Sum(nil)
	
	truncatedHash := hash[:15]
	encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(truncatedHash)
	
	var formattedKey strings.Builder
	for i, r := range encoded {
		if i > 0 && i%5 == 0 {
			formattedKey.WriteRune('-')
		}
		formattedKey.WriteRune(r)
	}
	return formattedKey.String()
}