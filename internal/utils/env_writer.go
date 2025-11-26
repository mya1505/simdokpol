package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync" // <-- Tambah ini
)

// Global Mutex untuk memastikan hanya 1 proses yang menulis .env dalam satu waktu
var envMutex sync.Mutex 

// UpdateEnvFile memperbarui key tertentu di file .env tanpa menghapus key lain.
func UpdateEnvFile(updates map[string]string) error {
	// --- FIX: LOCKING ---
	envMutex.Lock()
	defer envMutex.Unlock()
	// --------------------

	envPath := filepath.Join(GetAppDataDir(), ".env")

	// 1. Baca konten file yang ada
	content, err := os.ReadFile(envPath)
	if err != nil {
		// Jika file tidak ada, buat file baru jika perlu, atau return error
		// Disini kita return error agar explicit
		return fmt.Errorf("gagal membaca file .env: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	newLines := make([]string, 0, len(lines))
	processedKeys := make(map[string]bool)

	// 2. Update baris yang sudah ada
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "#") {
			newLines = append(newLines, line) // Simpan komentar & baris kosong
			continue
		}

		parts := strings.SplitN(trimmedLine, "=", 2)
		if len(parts) != 2 {
			newLines = append(newLines, line)
			continue
		}

		key := strings.TrimSpace(parts[0])
		if newValue, exists := updates[key]; exists {
			newLines = append(newLines, fmt.Sprintf("%s=%s", key, newValue)) //
			processedKeys[key] = true
		} else {
			newLines = append(newLines, line) // Simpan key lain yang tidak diubah
		}
	}

	// 3. Tambahkan key baru yang belum ada di file
	for key, value := range updates {
		if !processedKeys[key] {
			newLines = append(newLines, fmt.Sprintf("%s=%s", key, value)) //
		}
	}

	// 4. Tulis kembali ke file
	output := strings.Join(newLines, "\n")
	// Pastikan diakhiri newline
	if !strings.HasSuffix(output, "\n") {
		output += "\n"
	}

	// Gunakan permission 0600 (Read/Write only by owner) untuk keamanan
	if err := os.WriteFile(envPath, []byte(output), 0600); err != nil { //
		return fmt.Errorf("gagal menulis file .env: %w", err)
	}

	return nil
}
