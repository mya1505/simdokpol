package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"log"
	"os"
)

// Ini adalah tool helper HANYA UNTUK ANDA (DEVELOPER)
// Jalankan ini sekali untuk membuat private.pem (RAHASIA) dan public.pem (untuk app)

func main() {
	log.Println("Membuat ECDSA P-256 key pair...")
	// 1. Buat private key
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatalf("Gagal membuat private key: %v", err)
	}

	// 2. Simpan Private Key (private.pem)
	// KONVERSI KE FORMAT PKCS#8
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		log.Fatalf("Gagal marshal private key: %v", err)
	}
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	})
	if err := os.WriteFile("private.pem", privateKeyPEM, 0600); err != nil {
		log.Fatalf("Gagal menyimpan private.pem: %v", err)
	}
	log.Println("✅ Berhasil disimpan: private.pem (RAHASIA, JANGAN DIBAGIKAN!)")


	// 3. Simpan Public Key (public.pem)
	// KONVERSI KE FORMAT PKIX (Standar public key)
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		log.Fatalf("Gagal marshal public key: %v", err)
	}
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})
	if err := os.WriteFile("public.pem", publicKeyPEM, 0644); err != nil {
		log.Fatalf("Gagal menyimpan public.pem: %v", err)
	}
	log.Println("✅ Berhasil disimpan: public.pem (Tanam ini di dalam aplikasi)")
}