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

	caCertFile := filepath.Join(certDir, "ca.crt")
	caKeyFile := filepath.Join(certDir, "ca.key")
	certFile := filepath.Join(certDir, "server.crt")
	keyFile := filepath.Join(certDir, "server.key")

	if _, err := os.Stat(caCertFile); err == nil {
		if _, err := os.Stat(caKeyFile); err == nil {
			if _, err := os.Stat(certFile); err == nil {
				if _, err := os.Stat(keyFile); err == nil {
					return certFile, keyFile, nil
				}
			}
		}
	}

	return generateCAAndServerCert(caCertFile, caKeyFile, certFile, keyFile)
}

func generateCAAndServerCert(caCertPath, caKeyPath, serverCertPath, serverKeyPath string) (string, string, error) {
	// Generate CA Private Key
	caPriv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", err
	}

	// CA Template
	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour * 10) // 10 Tahun
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, _ := rand.Int(rand.Reader, serialNumberLimit)

	caTemplate := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"SIMDOKPOL Self-Signed"},
			CommonName:   "SIMDOKPOL Local CA",
		},
		NotBefore: notBefore, NotAfter: notAfter,

		IsCA:                  true,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	caDerBytes, err := x509.CreateCertificate(rand.Reader, &caTemplate, &caTemplate, &caPriv.PublicKey, caPriv)
	if err != nil {
		return "", "", err
	}

	// Save CA Cert
	caOut, _ := os.Create(caCertPath)
	pem.Encode(caOut, &pem.Block{Type: "CERTIFICATE", Bytes: caDerBytes})
	caOut.Close()

	// Save CA Key
	caKeyOut, _ := os.OpenFile(caKeyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	pem.Encode(caKeyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(caPriv)})
	caKeyOut.Close()

	// Generate Server Private Key
	serverPriv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", err
	}

	serverSerial, _ := rand.Int(rand.Reader, serialNumberLimit)
	serverTemplate := x509.Certificate{
		SerialNumber: serverSerial,
		Subject: pkix.Name{
			Organization: []string{"SIMDOKPOL Local Server"},
			CommonName:   "simdokpol.local",
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		IsCA:        false,
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},

		DNSNames:    []string{"localhost", "simdokpol.local"},
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
	}

	serverDerBytes, err := x509.CreateCertificate(rand.Reader, &serverTemplate, &caTemplate, &serverPriv.PublicKey, caPriv)
	if err != nil {
		return "", "", err
	}

	// Save Server Cert (include CA chain)
	serverOut, _ := os.Create(serverCertPath)
	pem.Encode(serverOut, &pem.Block{Type: "CERTIFICATE", Bytes: serverDerBytes})
	pem.Encode(serverOut, &pem.Block{Type: "CERTIFICATE", Bytes: caDerBytes})
	serverOut.Close()

	// Save Server Key
	serverKeyOut, _ := os.OpenFile(serverKeyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	pem.Encode(serverKeyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(serverPriv)})
	serverKeyOut.Close()

	return serverCertPath, serverKeyPath, nil
}
