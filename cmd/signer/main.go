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

// PASTIKAN KEY INI SAMA PERSIS DENGAN DI license_service.go
const appSecretKey = "GANTI_STRING_INI_DENGAN_KATA_SANDI_RAHASIA_YANG_PANJANG_DAN_RUMIT_12345"

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println(strings.Repeat("=", 50))
	fmt.Println("SIMDOKPOL KEY GENERATOR (HMAC)")
	fmt.Println(strings.Repeat("=", 50))

	// 1. Minta Input HWID
	fmt.Print("Masukkan Hardware ID dari User (misal: A1B2-C3D4-...): ")
	hwid, _ := reader.ReadString('\n')
	hwid = strings.TrimSpace(hwid)

	if hwid == "" {
		fmt.Println("Error: Hardware ID tidak boleh kosong.")
		return
	}

	// 2. Generate Key
	// Logic harus sama persis dengan generateSignature di license_service.go
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
	fmt.Println("SERIAL KEY GENERATED:")
	fmt.Println(strings.Repeat("-", 50))
	fmt.Printf("HWID Target : %s\n", hwid)
	fmt.Printf("Serial Key  : %s\n", finalKey)
	fmt.Println(strings.Repeat("=", 50))
}