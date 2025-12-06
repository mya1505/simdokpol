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

// Key default untuk Development (WAJIB SAMA dengan di service)
var appSecretKey = "JANGAN_PAKAI_DEFAULT_KEY_INI_BAHAYA"
//9b9f3e4e7142eb69ba8c68a33b16924fdfb46a5f3da44721a2502889b254b48d
func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println(strings.Repeat("=", 50))
	fmt.Println("   SIMDOKPOL KEY GENERATOR (DEV ONLY)")
	fmt.Println(strings.Repeat("=", 50))

	fmt.Print("👉 Masukkan Hardware ID User: ")
	hwid, _ := reader.ReadString('\n')
	hwid = strings.TrimSpace(hwid)

	if hwid == "" {
		fmt.Println("❌ Error: Hardware ID Kosong.")
		return
	}

	// Generate Logic
	h := hmac.New(sha256.New, []byte(appSecretKey))
	h.Write([]byte(hwid))
	hash := h.Sum(nil)

	truncatedHash := hash[:15]
	rawKey := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(truncatedHash)

	// Format Output
	var formattedKey strings.Builder
	for i, r := range rawKey {
		if i > 0 && i%5 == 0 {
			formattedKey.WriteRune('-')
		}
		formattedKey.WriteRune(r)
	}

	fmt.Println("\n✅ SERIAL KEY (DEV):")
	fmt.Println("--------------------------------------------------")
	fmt.Println(formattedKey.String())
	fmt.Println("--------------------------------------------------")
}