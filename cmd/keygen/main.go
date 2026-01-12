package main

import (
	"bufio"
	"crypto/ecdsa"
	"crypto/sha256"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"simdokpol/internal/utils"
)

func main() {
	// 1. PARSE FLAGS
	privateKeyPath := flag.String("key", "", "Path private key PEM (opsional)")
	flag.Parse()

	keyPath := resolvePrivateKeyPath(*privateKeyPath)
	privateKey, keyHash := loadPrivateKey(keyPath)

	reader := bufio.NewReader(os.Stdin)

	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("   SIMDOKPOL KEY GENERATOR (OFFLINE)")
	fmt.Println(strings.Repeat("=", 60))
	// Tampilkan fingerprint key biar Admin yakin ini key yang benar
	fmt.Printf("📁 Private Key    : %s\n", keyPath)
	fmt.Printf("🔑 Key Checksum   : %s...\n", keyHash)
	fmt.Println(strings.Repeat("-", 60))

	fmt.Print("👉 Masukkan Hardware ID User: ")
	hwid, _ := reader.ReadString('\n')
	hwid = strings.TrimSpace(hwid)

	if hwid == "" {
		fmt.Println("❌ Error: Hardware ID Kosong.")
		return
	}

	formattedKey, err := utils.SignActivationKey(hwid, privateKey)
	if err != nil {
		fmt.Printf("❌ Error: Gagal generate key: %v\n", err)
		return
	}

	fmt.Println("\n✅ SERIAL KEY VALID:")
	fmt.Println("--------------------------------------------------")
	fmt.Println(formattedKey)
	fmt.Println("--------------------------------------------------")
}

func sha256Sum(b []byte) string {
	h := sha256.Sum256(b)
	return fmt.Sprintf("%x", h[:8])
}

func resolvePrivateKeyPath(flagPath string) string {
	if flagPath != "" {
		return flagPath
	}

	if envPath := os.Getenv("LICENSE_PRIVATE_KEY"); envPath != "" {
		return envPath
	}

	if _, err := os.Stat("private.pem"); err == nil {
		return "private.pem"
	}

	return filepath.Join(utils.GetAppDataDir(), "private.pem")
}

func loadPrivateKey(path string) (*ecdsa.PrivateKey, string) {
	pemBytes, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("❌ CRITICAL ERROR: Gagal membaca private key: %v\n", err)
		os.Exit(1)
	}

	privateKey, err := utils.ParsePrivateKeyPEM(pemBytes)
	if err != nil {
		fmt.Printf("❌ CRITICAL ERROR: Private key tidak valid: %v\n", err)
		os.Exit(1)
	}

	return privateKey, sha256Sum(pemBytes)
}
