package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

func EnsureCertificates() (string, string, error) {
	certDir := filepath.Join(GetAppDataDir(), "certs")
	if err := os.MkdirAll(certDir, 0755); err != nil {
		return "", "", fmt.Errorf("gagal folder certs: %w", err)
	}

	certFile := filepath.Join(certDir, "server.crt")
	keyFile := filepath.Join(certDir, "server.key")

	if _, err := os.Stat(certFile); err == nil {
		if _, err := os.Stat(keyFile); err == nil {
			return certFile, keyFile, nil
		}
	}

	return generateSelfSignedCert(certFile, keyFile)
}

func generateSelfSignedCert(certPath, keyPath string) (string, string, error) {
	// Generate Private Key
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil { return "", "", err }

	// Setup Template
	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour * 10) // 10 Tahun
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, _ := rand.Int(rand.Reader, serialNumberLimit)

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"SIMDOKPOL Self-Signed"},
			CommonName:   "SIMDOKPOL Root CA", // Ganti nama jadi Root CA
		},
		NotBefore: notBefore, NotAfter: notAfter,
		
		// --- FIX: JADIKAN SEBAGAI CA (CERTIFICATE AUTHORITY) ---
		IsCA:                  true, // <--- INI KUNCINYA
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		
		DNSNames:    []string{"localhost", "simdokpol.local"},
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
	}

	// Create Certificate
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil { return "", "", err }

	// Save Cert
	certOut, _ := os.Create(certPath)
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()

	// Save Key
	keyOut, _ := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	keyOut.Close()

	return certPath, keyPath, nil
}