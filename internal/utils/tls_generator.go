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
	"os/exec" // <-- Penting untuk command sistem
	"path/filepath"
	"runtime"
	"time"
)

// EnsureCertificates memastikan file sertifikat ada. Jika tidak, buat baru.
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

// generateSelfSignedCert membuat self-signed certificate (valid 10 tahun)
func generateSelfSignedCert(certPath, keyPath string) (string, string, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", err
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour * 10) // 10 Tahun
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, _ := rand.Int(rand.Reader, serialNumberLimit)

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"SIMDOKPOL Self-Signed"},
			CommonName:   "SIMDOKPOL Local Server",
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,
		KeyUsage:  x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
		},
		BasicConstraintsValid: true,
		// Tambahkan domain lokal agar valid saat diakses
		DNSNames:    []string{"localhost", "simdokpol.local"},
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return "", "", err
	}

	certOut, _ := os.Create(certPath)
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()

	keyOut, _ := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	keyOut.Close()

	return certPath, keyPath, nil
}

// InstallCertToSystem (BARU): Meminta Windows untuk menginstal sertifikat ke Root Store
func InstallCertToSystem() error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("fitur auto-install sertifikat hanya tersedia di Windows")
	}

	certPath := filepath.Join(GetAppDataDir(), "certs", "server.crt")

	// Cek apakah file ada
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		return fmt.Errorf("file sertifikat belum dibuat")
	}

	// KONSEP:
	// Kita menggunakan PowerShell untuk memanggil 'Start-Process' dengan verb 'RunAs'.
	// 'RunAs' akan memicu popup UAC (Run as Administrator) yang wajib untuk memodifikasi Root Store.
	// Perintah intinya adalah: certutil -addstore -f "Root" "path\to\cert.crt"

	// Hati-hati dengan escaping path yang mungkin mengandung spasi
	cmdArgs := fmt.Sprintf("'-addstore', '-f', 'Root', '%s'", certPath)

	cmd := exec.Command("powershell", "Start-Process", "certutil",
		"-ArgumentList", cmdArgs,
		"-Verb", "RunAs", // Memicu UAC Prompt
		"-WindowStyle", "Hidden",
		"-Wait") // Tunggu sampai user klik Yes/No

	// Jalankan perintah
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gagal instalasi sertifikat (mungkin user menolak UAC): %w", err)
	}

	return nil
}