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

// Helper untuk parsing INT dengan default value (Fix Silent Failure)
func getEnvAsInt(key string, defaultVal int) int {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		log.Printf("⚠️ Warning: Invalid integer for ENV %s: %s. Using default: %d", key, valStr, defaultVal)
		return defaultVal
	}
	return val
}

func LoadConfig() *Config {
	appDataDir := utils.GetAppDataDir()
	envPath := filepath.Join(appDataDir, ".env")
	
	_ = godotenv.Load(envPath)
	_ = godotenv.Load()

	secretStr := os.Getenv("JWT_SECRET_KEY")
	if secretStr == "" {
		if os.Getenv("APP_ENV") == "production" {
			log.Println("⚠️ CRITICAL: Using default JWT Secret in Production!")
		}
		secretStr = "default-insecure-secret-change-me-immediately"
	}

	cost := getEnvAsInt("BCRYPT_COST", 10)
	if cost < bcrypt.MinCost {
		cost = bcrypt.MinCost
	}

	dialect := strings.ToLower(os.Getenv("DB_DIALECT"))
	if dialect == "" { dialect = "sqlite" }

	dsn := os.Getenv("DB_DSN")
	if dialect == "sqlite" {
		if dsn == "" {
			dsn = "simdokpol.db?_foreign_keys=on"
		}
		// Fix Path SQLite agar selalu absolute ke AppData
		if !strings.HasPrefix(dsn, "file:") && !filepath.IsAbs(strings.Split(dsn, "?")[0]) {
			dsn = filepath.Join(appDataDir, dsn)
		}
	}

	// Fix Env Key Case Sensitive di Linux
	archiveDays := getEnvAsInt("ARCHIVE_DURATION_DAYS", 15)
	if archiveDays == 0 {
		archiveDays = getEnvAsInt("archive_duration_days", 15)
	}

	return &Config{
		AppConfig: &dto.AppConfig{
			DBDialect:           dialect,
			DBDSN:               dsn,
			DBHost:              os.Getenv("DB_HOST"),
			DBPort:              os.Getenv("DB_PORT"),
			DBUser:              os.Getenv("DB_USER"),
			DBName:              os.Getenv("DB_NAME"),
			DBSSLMode:           os.Getenv("DB_SSLMODE"),
			IsSetupComplete:     os.Getenv("IS_SETUP_COMPLETE") == "true",
			
			KopBaris1:           os.Getenv("KOP_BARIS_1"),
			KopBaris2:           os.Getenv("KOP_BARIS_2"),
			KopBaris3:           os.Getenv("KOP_BARIS_3"),
			NamaKantor:          os.Getenv("NAMA_KANTOR"),
			TempatSurat:         os.Getenv("TEMPAT_SURAT"),
			FormatNomorSurat:    os.Getenv("FORMAT_NOMOR_SURAT"),
			NomorSuratTerakhir:  os.Getenv("NOMOR_SURAT_TERAKHIR"),
			ZonaWaktu:           os.Getenv("ZONA_WAKTU"),
			BackupPath:          os.Getenv("BACKUP_PATH"),
			
			ArchiveDurationDays: archiveDays,
			EnableHTTPS:         os.Getenv("ENABLE_HTTPS") == "true",
			SessionTimeout:      getEnvAsInt("SESSION_TIMEOUT", 480),
			IdleTimeout:         getEnvAsInt("IDLE_TIMEOUT", 15),
		},
		JWTSecretKey: []byte(secretStr),
		BcryptCost:   cost,
		DBPass:       os.Getenv("DB_PASS"),
	}
}