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
	"time"

	"simdokpol/internal/utils"

	"github.com/joho/godotenv"
)

var appSecretKey = ""

func main() {
	// 1. SETUP ENV
	envPath := filepath.Join(utils.GetAppDataDir(), ".env")
	_ = godotenv.Overload(envPath)

	hwidPtr := flag.String("id", "", "Hardware ID User")
	secretPtr := flag.String("secret", "", "Secret Key Manual (Override)")
	flag.Parse()

	// 2. RESOLUSI KEY
	if *secretPtr != "" {
		appSecretKey = *secretPtr
	}
	if appSecretKey == "" {
		appSecretKey = os.Getenv("APP_SECRET_KEY")
	}

	if appSecretKey == "" {
		fmt.Println("❌ CRITICAL ERROR: Secret Key belum dikonfigurasi!")
		fmt.Printf("   File .env di '%s' tidak memiliki APP_SECRET_KEY.\n", envPath)
		os.Exit(1)
	}

	targetHWID := *hwidPtr
	reader := bufio.NewReader(os.Stdin)

	if targetHWID == "" {
		fmt.Println(strings.Repeat("=", 60))
		fmt.Println("   SIMDOKPOL SIGNER CLI (SECURE)")
		fmt.Println("   " + time.Now().Format("2006-01-02 15:04:05"))
		fmt.Printf("   📁 Config: %s\n", envPath)
		fmt.Printf("   🔑 Key Hash: %s...\n", sha256Sum(appSecretKey))
		fmt.Println(strings.Repeat("=", 60))
		
		fmt.Print("🔧 Masukkan Hardware ID: ")
		input, _ := reader.ReadString('\n')
		targetHWID = strings.TrimSpace(input)
	}

	if targetHWID == "" {
		fmt.Println("❌ Error: HWID kosong.")
		return
	}

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

	if *hwidPtr == "" {
		fmt.Println("\nTekan Enter untuk keluar...")
		reader.ReadString('\n')
	}
}

func sha256Sum(s string) string {
	h := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", h[:8])
}