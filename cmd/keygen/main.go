package main

import (
	"bufio"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base32"
	"flag"
	"fmt"
	"os"
	"strings"
)

// Variable ini KOSONG secara default.
// Harus diisi saat build menggunakan: -ldflags "-X main.appSecretKey=RAHASIA_ASLI"
var appSecretKey = ""

func main() {
	// Fitur tambahan: Bisa input secret key via flag saat runtime (untuk admin)
	secretFlag := flag.String("secret", "", "App Secret Key manual (opsional)")
	flag.Parse()

	// Prioritas 1: Flag CLI
	if *secretFlag != "" {
		appSecretKey = *secretFlag
	}

	// Prioritas 2: Environment Variable
	if appSecretKey == "" {
		appSecretKey = os.Getenv("APP_SECRET_KEY")
	}

	// Validasi Safety
	if appSecretKey == "" {
		fmt.Println("❌ CRITICAL ERROR: Secret Key belum dikonfigurasi!")
		fmt.Println("Gunakan salah satu cara:")
		fmt.Println("1. Build dengan LDFLAGS: go build -ldflags \"-X main.appSecretKey=KUNCI_RAHASIA\"")
		fmt.Println("2. Jalankan dengan flag: ./keygen -secret=\"KUNCI_RAHASIA\"")
		os.Exit(1)
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Println(strings.Repeat("=", 50))
	fmt.Println("   SIMDOKPOL KEY GENERATOR (ADMIN TOOL)")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println("🔑 Secret Key Hash:", sha256Sum(appSecretKey)) // Print hashnya aja buat verifikasi (jangan key aslinya)
	fmt.Println(strings.Repeat("-", 50))

	fmt.Print("👉 Masukkan Hardware ID User: ")
	hwid, _ := reader.ReadString('\n')
	hwid = strings.TrimSpace(hwid)

	if hwid == "" {
		fmt.Println("❌ Error: Hardware ID Kosong.")
		return
	}

	// Generate Logic (SAMA PERSIS DENGAN LICENSE SERVICE)
	h := hmac.New(sha256.New, []byte(appSecretKey))
	h.Write([]byte(hwid))
	hash := h.Sum(nil)

	truncatedHash := hash[:15]
	rawKey := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(truncatedHash)

	// Format Output (XXXXX-XXXXX-XXXXX)
	var formattedKey strings.Builder
	for i, r := range rawKey {
		if i > 0 && i%5 == 0 {
			formattedKey.WriteRune('-')
		}
		formattedKey.WriteRune(r)
	}

	fmt.Println("\n✅ SERIAL KEY VALID:")
	fmt.Println("--------------------------------------------------")
	fmt.Println(formattedKey.String())
	fmt.Println("--------------------------------------------------")
	fmt.Println("💡 Copy key di atas dan berikan ke user.")
}

// Helper untuk print fingerprint secret key (biar admin tau dia pake key yg bener)
func sha256Sum(s string) string {
	h := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", h[:4]) + "..." // Cuma ambil 8 karakter awal
}