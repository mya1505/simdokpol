package config

import (
	"log"
	"os"
	"path/filepath"
	"simdokpol/internal/dto"
	"simdokpol/internal/utils"
	"strconv"
	"strings" // Pastikan import strings

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

// Config struct menampung konfigurasi kritikal startup
type Config struct {
	*dto.AppConfig
	JWTSecretKey []byte
	BcryptCost   int
	DBPass       string // <-- TAMBAHAN: Field khusus internal, tidak terekspos di JSON
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
			log.Fatal("FATAL: JWT_SECRET_KEY wajib diisi di mode production!")
		}
		log.Println("WARNING: Menggunakan default JWT Secret (TIDAK AMAN UNTUK PRODUCTION)")
		secretStr = "default-insecure-secret-change-me-immediately"
	}

	// 3. Determine Bcrypt Cost
	costStr := os.Getenv("BCRYPT_COST")
	cost, err := strconv.Atoi(costStr)
	if err != nil || cost < bcrypt.MinCost {
		cost = 10 // Default safe
	}
	
	// Logic SSL Mode (Default disable)
	sslMode := os.Getenv("DB_SSLMODE")
	if sslMode == "" { sslMode = "disable" }
	
	// Logic Dialect (Default sqlite)
	dialect := strings.ToLower(os.Getenv("DB_DIALECT"))
	if dialect == "" { dialect = "sqlite" }

	// 4. Return Config Object
	return &Config{
		AppConfig: &dto.AppConfig{
			DBDialect:           dialect,
			DBDSN:               os.Getenv("DB_DSN"),
			DBHost:              os.Getenv("DB_HOST"),
			DBPort:              os.Getenv("DB_PORT"),
			DBUser:              os.Getenv("DB_USER"),
			DBName:              os.Getenv("DB_NAME"),
			DBSSLMode:           sslMode,
			IsSetupComplete:     false, // Default, akan di-override service nanti
		},
		JWTSecretKey: []byte(secretStr),
		BcryptCost:   cost,
		DBPass:       os.Getenv("DB_PASS"), // <-- ISI DARI ENV
	}
}