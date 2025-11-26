// internal/utils/app_dir.go
package utils

import (
	"log"
	"os"
	"path/filepath"
)

var appDataDir string

// GetAppDataDir mengembalikan path ke direktori data aplikasi yang writeable.
// (e.g., %APPDATA%\SIMDOKPOL di Windows, ~/.config/simdokpol di Linux)
func GetAppDataDir() string {
	if appDataDir != "" {
		return appDataDir
	}

	// os.UserConfigDir() adalah lokasi yang tepat
	// Windows: C:\Users\<Nama>\AppData\Roaming
	// Linux: /home/<nama>/.config
	// macOS: /Users/<nama>/Library/Application Support
	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Printf("PERINGATAN: Gagal mendapatkan UserConfigDir, menggunakan direktori lokal: %v", err)
		// Fallback ke direktori saat ini jika gagal
		return "."
	}

	// Buat folder spesifik untuk aplikasi kita di dalamnya
	appSpecificDir := filepath.Join(configDir, "SIMDOKPOL")

	// Pastikan direktori ini ada
	if err := os.MkdirAll(appSpecificDir, 0755); err != nil {
		log.Printf("PERINGATAN: Gagal membuat direktori data aplikasi di %s, menggunakan direktori lokal: %v", appSpecificDir, err)
		// Fallback ke direktori saat ini jika gagal
		return "."
	}

	appDataDir = appSpecificDir
	log.Printf("INFO: Menggunakan direktori data aplikasi: %s", appDataDir)
	return appDataDir
}