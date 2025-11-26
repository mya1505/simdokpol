package utils

import (
	"encoding/base64"
)

// DecodeBase64 mendecode string base64 ke string biasa.
// Digunakan untuk memproses changelog yang di-inject saat build.
func DecodeBase64(s string) (string, error) {
	if s == "" {
		return "", nil
	}
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}