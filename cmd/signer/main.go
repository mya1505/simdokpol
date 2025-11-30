package main

import (
	"bufio"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base32"
	"fmt"
	"os"
	"strings"
	"time"
)

// === KONFIGURASI KEAMANAN ===
// Untuk production, gunakan -ldflags untuk inject key saat compile
// go build -ldflags="-X main.appSecretKey=YOUR_ACTUAL_SECRET_KEY" -o keygen.exe

var appSecretKey = "JANGAN_PAKAI_DEFAULT_KEY_INI_BAHAYA"

// === CONSTANTS ===
const (
	Version    = "2.0"
	HashLength = 15 // 15 bytes = 24 karakter base32
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("   SIMDOKPOL SECURE KEY GENERATOR v" + Version)
	fmt.Println("   " + time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println(strings.Repeat("=", 60))
	
	// Security Warning
	if isDefaultKey() {
		fmt.Println("⚠️  PERINGATAN KEAMANAN: Menggunakan kunci default!")
		fmt.Println("   Gunakan -ldflags untuk inject kunci rahasia saat compile")
		fmt.Println(strings.Repeat("-", 60))
	}

	// 1. Minta Input HWID
	fmt.Print("🔧 Masukkan Hardware ID User: ")
	hwid, _ := reader.ReadString('\n')
	hwid = strings.TrimSpace(hwid)

	if hwid == "" {
		fmt.Println("❌ Error: Hardware ID tidak boleh kosong.")
		exitPrompt(reader)
		return
	}

	// Validasi format HWID minimal
	if len(hwid) < 8 {
		fmt.Println("❌ Error: Hardware ID terlalu pendek.")
		exitPrompt(reader)
		return
	}

	// 2. Generate Secure Key
	serialKey, rawKey, err := generateSecureKey(hwid)
	if err != nil {
		fmt.Printf("❌ Error generating key: %v\n", err)
		exitPrompt(reader)
		return
	}

	// 3. Tampilkan Hasil
	fmt.Println("\n" + strings.Repeat("🔒", 30))
	fmt.Println("✅ SERIAL KEY BERHASIL DIBUAT")
	fmt.Println(strings.Repeat("🔒", 30))
	fmt.Printf("Tanggal    : %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("HWID Target: %s\n", hwid)
	fmt.Printf("Serial Key : %s\n", serialKey)
	
	if isDefaultKey() {
		fmt.Printf("Raw Key    : %s (DEBUG - HANYA UNTUK TESTING)\n", rawKey)
	}
	
	fmt.Println(strings.Repeat("=", 60))
	
	// 4. Security Notes
	fmt.Println("\n📝 CATATAN KEAMANAN:")
	fmt.Println("   • Simpan key ini di tempat yang aman")
	fmt.Println("   • Jangan bagikan file keygen.exe ke publik")
	fmt.Println("   • Setiap HWID menghasilkan key yang berbeda")
	fmt.Println("   • Key hanya berlaku untuk 1 device")
    
	exitPrompt(reader)
}

// generateSecureKey membuat key dengan HMAC-SHA256 yang aman
func generateSecureKey(hwid string) (formattedKey, rawKey string, err error) {
	// Gunakan HMAC dengan secret key
	h := hmac.New(sha256.New, []byte(appSecretKey))
	h.Write([]byte(hwid))
	hash := h.Sum(nil)

	// Ambil 15 byte pertama untuk balance keamanan & usability
	truncatedHash := hash[:HashLength]
	
	// Encode ke Base32 tanpa padding
	rawKey = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(truncatedHash)

	// Format dengan dash setiap 5 karakter untuk readability
	var formattedBuilder strings.Builder
	for i, r := range rawKey {
		if i > 0 && i%5 == 0 {
			formattedBuilder.WriteRune('-')
		}
		formattedBuilder.WriteRune(r)
	}
	
	formattedKey = formattedBuilder.String()
	return formattedKey, rawKey, nil
}

// isDefaultKey mengecek apakah masih menggunakan key default
func isDefaultKey() bool {
	return appSecretKey == "JANGAN_PAKAI_DEFAULT_KEY_INI_BAHAYA"
}

// exitPrompt menunggu user menekan enter sebelum keluar
func exitPrompt(reader *bufio.Reader) {
	fmt.Println("\nTekan Enter untuk keluar...")
	reader.ReadString('\n')
}