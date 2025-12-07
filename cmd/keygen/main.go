package main

import (
	"bufio"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base32"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"simdokpol/internal/utils" // Import Utils untuk dapat path folder

	"github.com/joho/godotenv" // Import Dotenv
)

// Variabel ini KOSONG defaultnya.
var appSecretKey = ""

func main() {
	// 1. SETUP ENV (Baca file .env dari folder config)
	envPath := filepath.Join(utils.GetAppDataDir(), ".env")
	_ = godotenv.Overload(envPath)

	// 2. PARSE FLAGS
	secretFlag := flag.String("secret", "", "App Secret Key manual (opsional)")
	flag.Parse()

	// 3. RESOLUSI SECRET KEY (Prioritas: Flag -> Env -> LDFLAGS)
	if *secretFlag != "" {
		appSecretKey = *secretFlag
	}

	if appSecretKey == "" {
		// Coba ambil dari Environment (yang sudah di-load dari file .env)
		appSecretKey = os.Getenv("APP_SECRET_KEY")
	}

	// 4. VALIDASI AKHIR
	if appSecretKey == "" {
		fmt.Println("❌ CRITICAL ERROR: Secret Key tidak ditemukan!")
		fmt.Printf("   File .env di '%s' tidak memiliki APP_SECRET_KEY.\n", envPath)
		fmt.Println("   Solusi: Jalankan aplikasi utama (simdokpol) dulu minimal sekali untuk generate key.")
		os.Exit(1)
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("   SIMDOKPOL KEY GENERATOR (SMART MODE)")
	fmt.Println(strings.Repeat("=", 60))
	// Tampilkan fingerprint key biar Admin yakin ini key yang benar
	fmt.Printf("📁 Config Source : %s\n", envPath)
	fmt.Printf("🔑 Key Checksum  : %s...\n", sha256Sum(appSecretKey)) 
	fmt.Println(strings.Repeat("-", 60))

	fmt.Print("👉 Masukkan Hardware ID User: ")
	hwid, _ := reader.ReadString('\n')
	hwid = strings.TrimSpace(hwid)

	if hwid == "" {
		fmt.Println("❌ Error: Hardware ID Kosong.")
		return
	}

	// Generate Logic (HMAC-SHA256)
	h := hmac.New(sha256.New, []byte(appSecretKey))
	h.Write([]byte(hwid))
	hash := h.Sum(nil)

	truncatedHash := hash[:15]
	rawKey := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(truncatedHash)

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
}

func sha256Sum(s string) string {
	h := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", h[:8])
}