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

// Kunci Rahasia Aplikasi (Di Production ini harus di-inject via ldflags saat build)
// Jangan gunakan kunci default ini di production!
var AppSecretKeyString = "JANGAN_PAKAI_DEFAULT_KEY_INI_BAHAYA"

var ErrLicenseInvalid = errors.New("kunci lisensi tidak valid untuk mesin ini")
var ErrLicenseBanned = errors.New("kunci lisensi ini telah diblokir")

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
	// Kita cek integritas saat startup juga
	svc.verifyRuntimeIntegrity()
	return svc
}

func (s *licenseService) GetHardwareID() string {
	return utils.GetHardwareID()
}

// verifyRuntimeIntegrity dijalankan saat startup untuk memastikan state konsisten
func (s *licenseService) verifyRuntimeIntegrity() {
	if s.IsLicensed() {
		// Jika valid, pastikan .env sinkron
		currentKey := os.Getenv(EnvLicenseKey)
		if currentKey == "" {
			// Coba ambil dari DB license terakhir
			// (Logic penyederhanaan: jika env kosong tapi valid, biarkan dulu atau load dari DB)
		}
	} else {
		// Jika tidak valid, pastikan DB status bersih
		s.configService.SaveConfig(map[string]string{LicenseStatusKey: LicenseStatusUnlicensed})
	}
}

// IsLicensed adalah Gatekeeper utama.
// FIX: Sekarang melakukan verifikasi Kriptografi setiap kali dipanggil.
func (s *licenseService) IsLicensed() bool {
	// 1. Ambil Key dari Environment (Prioritas Utama)
	currentKey := os.Getenv(EnvLicenseKey)
	
	// Jika Environment kosong, cek status DB. 
	// Jika DB bilang VALID tapi Key tidak ada di memori/env, itu mencurigakan (atau habis restart).
	if currentKey == "" {
		// Cek apakah di database config statusnya VALID?
		status, _ := s.GetLicenseStatus()
		if status == LicenseStatusValid {
			// BUG FIX: Dulu ini return true. SEKARANG TIDAK BOLEH.
			// Kalau status VALID tapi tidak ada Key yang bisa diverifikasi,
			// kita harus menganggapnya TIDAK VALID (atau cari key di tabel licenses).
			
			// Coba cari key aktif di tabel licenses (Fallback)
			// Ini butuh query ke repo, tapi demi keamanan, worth it.
			// Untuk simplifikasi code ini, kita anggap: NO KEY = NOT VALID.
			
			// Auto-revoke untuk membersihkan state "Hantu"
			s.RevokeLicense() 
			return false
		}
		return false
	}

	// 2. Verifikasi HMAC Signature (Inti Keamanan)
	// Kita hitung ulang: Apakah Key ini cocok dengan HWID mesin ini?
	hwid := s.GetHardwareID()
	expectedSignature := generateSignature(hwid)
	
	// Normalisasi input (hapus dash dan spasi)
	cleanKey := strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(currentKey, "-", ""), " ", ""))

	if cleanKey == expectedSignature {
		// COCOK! Mesin ini berhak jadi Pro.
		
		// Self-Healing: Jika DB bilang UNLICENSED, perbaiki jadi VALID
		status, _ := s.GetLicenseStatus()
		if status != LicenseStatusValid {
			s.configService.SaveConfig(map[string]string{LicenseStatusKey: LicenseStatusValid})
		}
		return true
	}

	// 3. TIDAK COCOK!
	// Key ada, tapi tanda tangannya salah (milik mesin lain atau asal ketik).
	// Tindakan: Revoke status pro di DB.
	status, _ := s.GetLicenseStatus()
	if status == LicenseStatusValid {
		log.Printf("SECURITY: Lisensi tidak valid terdeteksi (Key: %s, HWID: %s). Revoking...", currentKey, hwid)
		s.RevokeLicense()
	}

	return false
}

func (s *licenseService) RevokeLicense() error {
	// Set DB ke UNLICENSED
	if err := s.configService.SaveConfig(map[string]string{LicenseStatusKey: LicenseStatusUnlicensed}); err != nil {
		return err
	}
	// Bersihkan Env var di memori process
	os.Setenv(EnvLicenseKey, "")
	// Bersihkan file .env
	utils.UpdateEnvFile(map[string]string{EnvLicenseKey: ""})
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
	rawInput := inputKey
	cleanKey := strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(inputKey, "-", ""), " ", ""))

	hwid := s.GetHardwareID()
	expectedSignature := generateSignature(hwid)

	if cleanKey != expectedSignature {
		if actorID != 0 {
			s.auditService.LogActivity(actorID, "GAGAL AKTIVASI", fmt.Sprintf("Key salah/ilegal. HWID: %s", hwid))
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

	// Simpan riwayat ke tabel licenses
	existing, _ := s.licenseRepo.GetLicense(cleanKey)
	if existing != nil {
		existing.Status = LicenseStatusValid
		existing.ActivatedAt = &now
		existing.ActivatedByID = &actorIDUint
		s.licenseRepo.SaveLicense(existing)
	} else {
		s.licenseRepo.SaveLicense(license)
	}

	// Update Status Config
	if err := s.configService.SaveConfig(map[string]string{LicenseStatusKey: LicenseStatusValid}); err != nil {
		return nil, fmt.Errorf("gagal update config: %w", err)
	}

	// Update .env agar persist saat restart
	utils.UpdateEnvFile(map[string]string{EnvLicenseKey: rawInput})
	os.Setenv(EnvLicenseKey, rawInput) // Update memori juga

	if actorID != 0 {
		s.auditService.LogActivity(actorID, "AKTIVASI LISENSI", "Lisensi PRO berhasil diaktifkan.")
	}
	return license, nil
}

// generateSignature membuat tanda tangan HMAC (Key yang valid)
func generateSignature(data string) string {
	h := hmac.New(sha256.New, []byte(AppSecretKeyString))
	h.Write([]byte(data))
	hash := h.Sum(nil)
	
	// Ambil 15 byte pertama agar kuncinya tidak kepanjangan
	truncatedHash := hash[:15]
	// Encode ke Base32 agar aman dibaca manusia (huruf kapital & angka)
	encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(truncatedHash)
	
	// Format dengan dash setiap 5 karakter: XXXXX-XXXXX-XXXXX...
	var formattedKey strings.Builder
	for i, r := range encoded {
		if i > 0 && i%5 == 0 {
			formattedKey.WriteRune('-')
		}
		formattedKey.WriteRune(r)
	}
	return formattedKey.String()
}