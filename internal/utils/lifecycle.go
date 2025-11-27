package utils

import (
	"log"
	"os"
	"os/exec"
)

// RestartApp me-restart aplikasi saat ini.
// Cocok digunakan setelah perubahan konfigurasi kritis (DB/HTTPS).
func RestartApp() error {
	// 1. Dapatkan path executable saat ini
	self, err := os.Executable()
	if err != nil {
		return err
	}

	// 2. Siapkan perintah untuk menjalankan ulang diri sendiri
	// os.Args[1:] meneruskan argumen CLI (jika ada) ke proses baru
	cmd := exec.Command(self, os.Args[1:]...)
	
	// Teruskan stdout/stderr agar log tetap jalan (untuk mode console)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	// Detach process (biar tidak ikut mati saat parent mati)
	// Di Windows/Linux Go biasanya handle ini otomatis via Start()
	
	// 3. Jalankan proses baru
	log.Println("🔄 SYSTEM RESTART TRIGGERED...")
	if err := cmd.Start(); err != nil {
		return err
	}

	// 4. Matikan proses saat ini
	os.Exit(0)
	return nil
}