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
	"time"
)

// Variabel ini KOSONG. Harus diisi saat build via LDFLAGS.
// go build -ldflags "-X main.appSecretKey=RAHASIA_ASLI"
var appSecretKey = ""

func main() {
	// Support args -id="..." dan -secret="..."
	hwidPtr := flag.String("id", "", "Hardware ID User")
	secretPtr := flag.String("secret", "", "Secret Key Manual (Override)")
	flag.Parse()

	// --- 1. RESOLUSI SECRET KEY ---
	// Urutan prioritas: Flag > Env > LDFLAGS (Variable)
	if *secretPtr != "" {
		appSecretKey = *secretPtr
	}
	if appSecretKey == "" {
		appSecretKey = os.Getenv("APP_SECRET_KEY")
	}

	// Validasi Safety
	if appSecretKey == "" {
		fmt.Println("❌ CRITICAL ERROR: Secret Key belum dikonfigurasi!")
		fmt.Println("Gunakan salah satu cara:")
		fmt.Println("1. Build: go build -ldflags \"-X main.appSecretKey=KEY_RAHASIA\"")
		fmt.Println("2. Run:   go run cmd/signer/main.go -secret=\"KEY_RAHASIA\"")
		os.Exit(1)
	}

	targetHWID := *hwidPtr
	reader := bufio.NewReader(os.Stdin)

	// --- 2. MODE INTERAKTIF ---
	if targetHWID == "" {
		fmt.Println(strings.Repeat("=", 60))
		fmt.Println("   SIMDOKPOL SIGNER CLI (SECURE)")
		fmt.Println("   " + time.Now().Format("2006-01-02 15:04:05"))
		// Tampilkan fingerprint key (4 karakter hash) untuk verifikasi visual admin
		fmt.Printf("   🔑 Key Checksum: %s...\n", sha256Sum(appSecretKey))
		fmt.Println(strings.Repeat("=", 60))
		
		fmt.Print("🔧 Masukkan Hardware ID: ")
		input, _ := reader.ReadString('\n')
		targetHWID = strings.TrimSpace(input)
	}

	if targetHWID == "" {
		fmt.Println("❌ Error: HWID kosong.")
		return
	}

	// --- 3. LOGIC GENERATE (HMAC-SHA256) ---
	h := hmac.New(sha256.New, []byte(appSecretKey))
	h.Write([]byte(targetHWID))
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

	fmt.Println("\n✅ SERIAL KEY:")
	fmt.Println("--------------------------------------------------")
	fmt.Println(formattedKey.String())
	fmt.Println("--------------------------------------------------")

	// Pause biar jendela console gak langsung nutup di Windows
	if *hwidPtr == "" {
		fmt.Println("\nTekan Enter untuk keluar...")
		reader.ReadString('\n')
	}
}

// Helper untuk menampilkan fingerprint key (bukan key aslinya)
func sha256Sum(s string) string {
	h := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", h[:4])
}