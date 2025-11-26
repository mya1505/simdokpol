package config

import (
	"log"
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
	// 1. Load .env dari AppData (Konsisten!)
	appDataDir := utils.GetAppDataDir()
	envPath := filepath.Join(appDataDir, ".env")
	
	// Coba load, ignore error kalau belum ada (first run)
	_ = godotenv.Load(envPath)

	// 2. Setup JWT
	secretStr := os.Getenv("JWT_SECRET_KEY")
	if secretStr == "" {
		secretStr = "default-insecure-secret-change-me"
	}

	// 3. Cost Bcrypt
	costStr := os.Getenv("BCRYPT_COST")
	cost, _ := strconv.Atoi(costStr)
	if cost < bcrypt.MinCost { cost = 10 }
	
	// 4. Database Logic (CRITICAL FIX)
	dialect := strings.ToLower(os.Getenv("DB_DIALECT"))
	if dialect == "" { dialect = "sqlite" }

	dsn := os.Getenv("DB_DSN")
	
	// JIKA SQLITE: Paksa Absolute Path ke AppData jika DSN masih relatif
	if dialect == "sqlite" {
		if dsn == "" {
			// Default DSN
			dsn = filepath.Join(appDataDir, "simdokpol.db?_foreign_keys=on")
		} else if !filepath.IsAbs(strings.Split(dsn, "?")[0]) {
			// Jika user set "simdokpol.db", kita ubah jadi "C:\Users\Name\AppData\...\simdokpol.db"
			cleanDSN := strings.TrimPrefix(dsn, "file:")
			// Ambil filename saja (ignore query params)
			parts := strings.Split(cleanDSN, "?")
			fname := filepath.Base(parts[0])
			
			// Rebuild DSN dengan path absolut
			dsn = filepath.Join(appDataDir, fname)
			if len(parts) > 1 {
				dsn += "?" + parts[1]
			}
		}
	}

	return &Config{
		AppConfig: &dto.AppConfig{
			DBDialect:           dialect,
			DBDSN:               dsn, // <-- DSN YANG SUDAH DIPERBAIKI
			DBHost:              os.Getenv("DB_HOST"),
			DBPort:              os.Getenv("DB_PORT"),
			DBUser:              os.Getenv("DB_USER"),
			DBName:              os.Getenv("DB_NAME"),
			DBSSLMode:           os.Getenv("DB_SSLMODE"),
			IsSetupComplete:     os.Getenv("is_setup_complete") == "true",
			
			// Load settings lainnya...
			KopBaris1:           os.Getenv("kop_baris_1"),
			KopBaris2:           os.Getenv("kop_baris_2"),
			KopBaris3:           os.Getenv("kop_baris_3"),
			NamaKantor:          os.Getenv("nama_kantor"),
			TempatSurat:         os.Getenv("tempat_surat"),
			FormatNomorSurat:    os.Getenv("format_nomor_surat"),
			NomorSuratTerakhir:  os.Getenv("nomor_surat_terakhir"),
			ZonaWaktu:           os.Getenv("zona_waktu"),
			ArchiveDurationDays: 15, // Default
		},
		JWTSecretKey: []byte(secretStr),
		BcryptCost:   cost,
		DBPass:       os.Getenv("DB_PASS"),
	}
}