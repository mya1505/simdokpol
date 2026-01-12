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
	"time"

	"simdokpol/internal/utils"
)

func main() {
	hwidPtr := flag.String("id", "", "Hardware ID User")
	privateKeyPath := flag.String("key", "", "Path private key PEM (opsional)")
	flag.Parse()

	keyPath := resolvePrivateKeyPath(*privateKeyPath)
	privateKey, keyHash := loadPrivateKey(keyPath)

	targetHWID := *hwidPtr
	reader := bufio.NewReader(os.Stdin)

	if targetHWID == "" {
		fmt.Println(strings.Repeat("=", 60))
		fmt.Println("   SIMDOKPOL SIGNER CLI (SECURE)")
		fmt.Println("   " + time.Now().Format("2006-01-02 15:04:05"))
		fmt.Printf("   📁 Private Key: %s\n", keyPath)
		fmt.Printf("   🔑 Key Hash: %s...\n", keyHash)
		fmt.Println(strings.Repeat("=", 60))

		fmt.Print("🔧 Masukkan Hardware ID: ")
		input, _ := reader.ReadString('\n')
		targetHWID = strings.TrimSpace(input)
	}

	if targetHWID == "" {
		fmt.Println("❌ Error: HWID kosong.")
		return
	}

	formattedKey, err := utils.SignActivationKey(targetHWID, privateKey)
	if err != nil {
		fmt.Printf("❌ Error: Gagal generate key: %v\n", err)
		return
	}

	fmt.Println("\n✅ SERIAL KEY:")
	fmt.Println("--------------------------------------------------")
	fmt.Println(formattedKey)
	fmt.Println("--------------------------------------------------")

	if *hwidPtr == "" {
		fmt.Println("\nTekan Enter untuk keluar...")
		reader.ReadString('\n')
	}
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
