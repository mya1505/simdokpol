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

// Variabel ini akan DI-TIMPA oleh Makefile saat build release
var appSecretKey = "JANGAN_PAKAI_DEFAULT_KEY_INI_BAHAYA"

func main() {
	// Support args -id="..."
	hwidPtr := flag.String("id", "", "Hardware ID User")
	flag.Parse()

	targetHWID := *hwidPtr
	reader := bufio.NewReader(os.Stdin)

	// Mode Interaktif jika arg kosong
	if targetHWID == "" {
		fmt.Println(strings.Repeat("=", 60))
		fmt.Println("   SIMDOKPOL SIGNER CLI (PRODUCTION)")
		fmt.Println("   " + time.Now().Format("2006-01-02 15:04:05"))
		fmt.Println(strings.Repeat("=", 60))
		
		fmt.Print("🔧 Masukkan Hardware ID: ")
		input, _ := reader.ReadString('\n')
		targetHWID = strings.TrimSpace(input)
	}

	if targetHWID == "" {
		fmt.Println("❌ Error: HWID kosong.")
		return
	}

	// Generate Logic
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