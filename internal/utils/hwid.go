package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"runtime"
	"strings"
)

// GetHardwareID menghasilkan string unik berdasarkan fingerprint mesin.
func GetHardwareID() string {
	// Kita gabungkan beberapa info unik mesin
	hostname, _ := os.Hostname()
	platform := runtime.GOOS + "/" + runtime.GOARCH
	cpus := runtime.NumCPU()

	// Buat raw string: "hostname|linux/amd64|8"
	rawData := fmt.Sprintf("%s|%s|%d", hostname, platform, cpus)

	// Hash data tersebut
	hash := sha256.Sum256([]byte(rawData))
	
	// Ambil 4 byte pertama dan terakhir untuk ID pendek (16 karakter hex)
	// Contoh: A1B2-C3D4-E5F6-G7H8
	fullHex := hex.EncodeToString(hash[:])
	
	// Format jadi 4 blok @ 4 karakter
	if len(fullHex) >= 16 {
		part1 := strings.ToUpper(fullHex[0:4])
		part2 := strings.ToUpper(fullHex[4:8])
		part3 := strings.ToUpper(fullHex[8:12])
		part4 := strings.ToUpper(fullHex[12:16])
		return fmt.Sprintf("%s-%s-%s-%s", part1, part2, part3, part4)
	}
	
	return "UNKNOWN-HWID"
}