package utils

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func loadTestPrivateKey(t *testing.T) *ecdsa.PrivateKey {
	t.Helper()

	paths := []string{
		"private.pem",
		filepath.Join("..", "..", "private.pem"),
	}

	for _, path := range paths {
		if pemBytes, err := os.ReadFile(path); err == nil {
			key, err := ParsePrivateKeyPEM(pemBytes)
			if err != nil {
				t.Fatalf("gagal parse private key: %v", err)
			}
			return key
		}
	}

	t.Fatal("private.pem tidak ditemukan untuk test")
	return nil
}

func TestVerifyActivationKeyBase64(t *testing.T) {
	privateKey := loadTestPrivateKey(t)
	hwid := "HWID-TEST-1234"

	key, err := SignActivationKey(hwid, privateKey)
	if err != nil {
		t.Fatalf("gagal sign activation key: %v", err)
	}

	if !VerifyActivationKey(hwid, key) {
		t.Fatal("activation key base64 gagal diverifikasi")
	}
}

func TestVerifyActivationKeyBase32Legacy(t *testing.T) {
	privateKey := loadTestPrivateKey(t)
	hwid := "HWID-TEST-LEGACY"

	hash := sha256.Sum256([]byte(hwid))
	signature, err := ecdsa.SignASN1(rand.Reader, privateKey, hash[:])
	if err != nil {
		t.Fatalf("gagal sign legacy key: %v", err)
	}

	raw := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(signature)
	legacyKey := strings.ToUpper(raw)

	if !VerifyActivationKey(hwid, legacyKey) {
		t.Fatal("activation key base32 gagal diverifikasi")
	}
}
