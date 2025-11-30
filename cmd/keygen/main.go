package main

import (
	"bufio"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base32"
	"fmt"
	"os"
	"strings"
)

// FIX: Gunakan 'var' agar bisa di-inject via -ldflags di Makefile
var appSecretKey = "JANGAN_PAKAI_DEFAULT_KEY_INI_BAHAYA"

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println(strings.Repeat("=", 50))
	fmt.Println("   SIMDOKPOL KEY GENERATOR (CLI)")
	fmt.Println(strings.Repeat("=", 50))

	// 1. Minta Input HWID
	fmt.Print("👉 Masukkan Hardware ID User: ")
	hwid, _ := reader.ReadString('\n')
	hwid = strings.TrimSpace(hwid)

	if hwid == "" {
		fmt.Println("❌ Error: Hardware ID tidak boleh kosong.")
		fmt.Println("   Tekan Enter untuk keluar...")
		reader.ReadString('\n')
		return
	}

	// 2. Generate Key (Logic Sama Persis dengan App Utama)
	h := hmac.New(sha256.New, []byte(appSecretKey))
	h.Write([]byte(hwid))
	hash := h.Sum(nil)

	truncatedHash := hash[:15] // Ambil 15 byte
	rawKey := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(truncatedHash)

	// 3. Format Key (XXXXX-XXXXX-...)
	var formattedKey strings.Builder
	for i, r := range rawKey {
		if i > 0 && i%5 == 0 {
			formattedKey.WriteRune('-')
		}
		formattedKey.WriteRune(r)
	}
	
	finalKey := formattedKey.String()

	fmt.Println("\n" + strings.Repeat("-", 50))
	fmt.Println("✅ SERIAL KEY VALID:")
	fmt.Println(strings.Repeat("-", 50))
	fmt.Printf("HWID Target : %s\n", hwid)
	fmt.Printf("Serial Key  : %s\n", finalKey)
	fmt.Printf("Raw Key     : %s\n", rawKey) // Untuk debug
	fmt.Println(strings.Repeat("=", 50))
    
    // Pause biar window gak langsung nutup di Windows
    fmt.Println("\nTekan Enter untuk keluar...")
    reader.ReadString('\n')
}