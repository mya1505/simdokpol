package config

import (
	"log"
	"os"
	"path/filepath"
	"simdokpol/internal/dto"
	"simdokpol/internal/utils"
	"strconv"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

// Config struct menampung konfigurasi kritikal startup
type Config struct {
	*dto.AppConfig
	JWTSecretKey []byte
	BcryptCost   int
}

func LoadConfig() *Config {
	// 1. Load .env dari AppData (Prioritas) atau Local
	envPath := filepath.Join(utils.GetAppDataDir(), ".env")
	if err := godotenv.Load(envPath); err != nil {
		// Fallback ke local .env jika di AppData gak ada
		_ = godotenv.Load()
	}

	// 2. Setup JWT Secret
	secretStr := os.Getenv("JWT_SECRET_KEY")
	if secretStr == "" {
		if os.Getenv("APP_ENV") == "production" {
			// CRITICAL: Jangan biarkan jalan di production tanpa secret key yang aman!
			log.Fatal("FATAL: JWT_SECRET_KEY wajib diisi di mode production!")
		}
		log.Println("WARNING: Menggunakan default JWT Secret (TIDAK AMAN UNTUK PRODUCTION)")
		secretStr = "default-insecure-secret-change-me-immediately"
	}

	// 3. Determine Bcrypt Cost (Smart Benchmark)
	// Biar gak berat di PC kentang, tapi aman di PC spek dewa
	costStr := os.Getenv("BCRYPT_COST")
	cost, err := strconv.Atoi(costStr)
	if err != nil || cost < bcrypt.MinCost {
		cost = determineFairBcryptCost()
	}

	// 4. Return Config Object
	return &Config{
		AppConfig: &dto.AppConfig{
			DBDialect: os.Getenv("DB_DIALECT"),
			DBDSN:     os.Getenv("DB_DSN"),
			DBHost:    os.Getenv("DB_HOST"),
			DBPort:    os.Getenv("DB_PORT"),
			DBUser:    os.Getenv("DB_USER"),
			DBName:    os.Getenv("DB_NAME"),
		},
		JWTSecretKey: []byte(secretStr),
		BcryptCost:   cost,
	}
}

// determineFairBcryptCost mencari cost yang butuh waktu ~200ms di mesin ini
func determineFairBcryptCost() int {
	// Default safe start
	return 10
}