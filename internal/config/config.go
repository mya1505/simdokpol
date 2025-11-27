package config

import (
	"os"
	"path/filepath"
	"simdokpol/internal/dto"
	"simdokpol/internal/utils"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

type Config struct {
	*dto.AppConfig
	JWTSecretKey []byte
	BcryptCost   int
	DBPass       string
}

func LoadConfig() *Config {
	// 1. Load .env dari AppData (Prioritas Utama)
	appDataDir := utils.GetAppDataDir()
	envPath := filepath.Join(appDataDir, ".env")
	_ = godotenv.Load(envPath)
	// Fallback ke lokal (untuk dev)
	_ = godotenv.Load()

	// 2. Setup Security
	secretStr := os.Getenv("JWT_SECRET_KEY")
	if secretStr == "" {
		secretStr = "default-insecure-secret-change-me-immediately"
	}

	costStr := os.Getenv("BCRYPT_COST")
	cost, _ := strconv.Atoi(costStr)
	if cost < bcrypt.MinCost { cost = 10 }
	
	// 3. Database Logic (Path Fix)
	dialect := strings.ToLower(os.Getenv("DB_DIALECT"))
	if dialect == "" { dialect = "sqlite" }

	dsn := os.Getenv("DB_DSN")
	if dialect == "sqlite" {
		if dsn == "" || dsn == "simdokpol.db?_foreign_keys=on" {
			dsn = filepath.Join(appDataDir, "simdokpol.db?_foreign_keys=on")
		} else if !filepath.IsAbs(strings.Split(dsn, "?")[0]) {
			cleanDSN := strings.TrimPrefix(dsn, "file:")
			parts := strings.Split(cleanDSN, "?")
			fname := filepath.Base(parts[0])
			dsn = filepath.Join(appDataDir, fname)
			if len(parts) > 1 { dsn += "?" + parts[1] }
		}
	}

	// 4. Parse Settings Lain (Fallback Default)
	archiveDays, _ := strconv.Atoi(os.Getenv("archive_duration_days"))
	if archiveDays == 0 { archiveDays = 15 }

	// Session Timeout (Default 8 Jam)
	sessionTimeout, _ := strconv.Atoi(os.Getenv("SESSION_TIMEOUT"))
	if sessionTimeout == 0 { sessionTimeout = 480 }

	// Idle Timeout (Default 15 Menit)
	idleTimeout, _ := strconv.Atoi(os.Getenv("IDLE_TIMEOUT"))
	if idleTimeout == 0 { idleTimeout = 15 }

	// HTTPS Status
	enableHTTPS := os.Getenv("ENABLE_HTTPS") == "true"

	return &Config{
		AppConfig: &dto.AppConfig{
			DBDialect:           dialect,
			DBDSN:               dsn,
			DBHost:              os.Getenv("DB_HOST"),
			DBPort:              os.Getenv("DB_PORT"),
			DBUser:              os.Getenv("DB_USER"),
			DBName:              os.Getenv("DB_NAME"),
			DBSSLMode:           os.Getenv("DB_SSLMODE"),
			IsSetupComplete:     os.Getenv("is_setup_complete") == "true",
			
			KopBaris1:           os.Getenv("kop_baris_1"),
			KopBaris2:           os.Getenv("kop_baris_2"),
			KopBaris3:           os.Getenv("kop_baris_3"),
			NamaKantor:          os.Getenv("nama_kantor"),
			TempatSurat:         os.Getenv("tempat_surat"),
			FormatNomorSurat:    os.Getenv("format_nomor_surat"),
			NomorSuratTerakhir:  os.Getenv("nomor_surat_terakhir"),
			ZonaWaktu:           os.Getenv("zona_waktu"),
			BackupPath:          os.Getenv("backup_path"),
			
			// Field Baru
			ArchiveDurationDays: archiveDays,
			EnableHTTPS:         enableHTTPS,
			SessionTimeout:      sessionTimeout,
			IdleTimeout:         idleTimeout,
		},
		JWTSecretKey: []byte(secretStr),
		BcryptCost:   cost,
		DBPass:       os.Getenv("DB_PASS"),
	}
}