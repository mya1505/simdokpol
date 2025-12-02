package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Global Mutex untuk thread-safety
var envMutex sync.Mutex

// UpdateEnvFile memperbarui key di .env.
// FIX: Sekarang otomatis membuat file jika belum ada.
func UpdateEnvFile(updates map[string]string) error {
	envMutex.Lock()
	defer envMutex.Unlock()

	envPath := filepath.Join(GetAppDataDir(), ".env")

	// 1. Baca konten file yang ada
	var lines []string
	content, err := os.ReadFile(envPath)
	if err != nil {
		if os.IsNotExist(err) {
			// FIX: Jika file tidak ada, jangan error. Anggap kosong & buat baru.
			lines = []string{}
		} else {
			// Jika error lain (misal permission denied), baru lapor.
			return fmt.Errorf("gagal membaca file .env: %w", err)
		}
	} else {
		lines = strings.Split(string(content), "\n")
	}

	newLines := make([]string, 0, len(lines)+len(updates))
	processedKeys := make(map[string]bool)

	// 2. Update baris yang sudah ada (Replace Value)
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		
		// Pertahankan baris kosong atau komentar
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "#") {
			newLines = append(newLines, line)
			continue
		}

		parts := strings.SplitN(trimmedLine, "=", 2)
		if len(parts) != 2 {
			newLines = append(newLines, line) // Baris rusak biarkan saja
			continue
		}

		key := strings.TrimSpace(parts[0])
		if newValue, exists := updates[key]; exists {
			// Ganti dengan nilai baru
			newLines = append(newLines, fmt.Sprintf("%s=%s", key, newValue))
			processedKeys[key] = true
		} else {
			// Pertahankan nilai lama
			newLines = append(newLines, line)
		}
	}

	// 3. Tambahkan key baru (Append New)
	for key, value := range updates {
		if !processedKeys[key] {
			newLines = append(newLines, fmt.Sprintf("%s=%s", key, value))
		}
	}

	// 4. Gabungkan & Tulis ke Disk
	output := strings.Join(newLines, "\n")
	// Pastikan file diakhiri baris baru (standar POSIX)
	if len(output) > 0 && !strings.HasSuffix(output, "\n") {
		output += "\n"
	}

	// WriteFile akan otomatis membuat file jika belum ada, atau menimpa jika ada.
	// Permission 0600 = Read/Write hanya untuk user pemilik (Secure)
	if err := os.WriteFile(envPath, []byte(output), 0600); err != nil {
		return fmt.Errorf("gagal menulis file .env: %w", err)
	}

	return nil
}