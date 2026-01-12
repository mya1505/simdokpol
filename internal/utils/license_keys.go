package utils

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base32"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"strings"
	"sync"
)

const licensePublicKeyPEM = `-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE6iw4qOiJ2++0HCctnlrKs0xiyqGl
bpxP8ee0jRx1F4UuaUWSO7aFWTl2OIrt9FeHgbRsTY/DDnYHamT9nJ+Tdw==
-----END PUBLIC KEY-----`

var (
	licensePublicKey     *ecdsa.PublicKey
	licensePublicKeyOnce sync.Once
	licensePublicKeyErr  error
)

func NormalizeActivationKey(input string) string {
	cleaned := strings.ReplaceAll(input, " ", "")
	if isBase32Key(cleaned) {
		return strings.ToUpper(strings.ReplaceAll(cleaned, "-", ""))
	}
	return cleaned
}

func FormatActivationKey(raw string) string {
	return NormalizeActivationKey(raw)
}

func DecodeActivationKey(input string) ([]byte, error) {
	cleaned := strings.ReplaceAll(input, " ", "")
	if cleaned == "" {
		return nil, errors.New("activation key kosong")
	}

	if isBase32Key(cleaned) {
		normalized := strings.ToUpper(strings.ReplaceAll(cleaned, "-", ""))
		return base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(normalized)
	}

	return base64.RawURLEncoding.DecodeString(cleaned)
}

func EncodeActivationKey(signature []byte) string {
	return base64.RawURLEncoding.EncodeToString(signature)
}

func VerifyActivationKey(hwid string, activationKey string) bool {
	signature, err := DecodeActivationKey(activationKey)
	if err != nil {
		return false
	}

	pub, err := getLicensePublicKey()
	if err != nil {
		return false
	}

	hash := sha256.Sum256([]byte(hwid))
	return ecdsa.VerifyASN1(pub, hash[:], signature)
}

func SignActivationKey(hwid string, privateKey *ecdsa.PrivateKey) (string, error) {
	if privateKey == nil {
		return "", errors.New("private key nil")
	}
	hash := sha256.Sum256([]byte(hwid))
	signature, err := ecdsa.SignASN1(rand.Reader, privateKey, hash[:])
	if err != nil {
		return "", err
	}
	return EncodeActivationKey(signature), nil
}

func isBase32Key(input string) bool {
	if input == "" {
		return false
	}

	for _, r := range input {
		switch {
		case r >= 'A' && r <= 'Z':
		case r >= 'a' && r <= 'z':
		case r >= '2' && r <= '7':
		case r == '-':
		default:
			return false
		}
	}
	return true
}

func ParsePrivateKeyPEM(pemBytes []byte) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("private key PEM tidak valid")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	ecdsaKey, ok := key.(*ecdsa.PrivateKey)
	if !ok {
		return nil, errors.New("private key bukan ECDSA")
	}
	return ecdsaKey, nil
}

func getLicensePublicKey() (*ecdsa.PublicKey, error) {
	licensePublicKeyOnce.Do(func() {
		block, _ := pem.Decode([]byte(licensePublicKeyPEM))
		if block == nil {
			licensePublicKeyErr = errors.New("public key PEM tidak valid")
			return
		}
		key, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			licensePublicKeyErr = err
			return
		}
		pubKey, ok := key.(*ecdsa.PublicKey)
		if !ok {
			licensePublicKeyErr = errors.New("public key bukan ECDSA")
			return
		}
		licensePublicKey = pubKey
	})
	return licensePublicKey, licensePublicKeyErr
}
