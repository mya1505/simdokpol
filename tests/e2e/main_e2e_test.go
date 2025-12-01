package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Konfigurasi Target Test
const baseURL = "http://localhost:8080"

// Struct untuk request login
type LoginReq struct {
	NRP      string `json:"nrp"`
	Password string `json:"password"`
}

// TestMainE2E: Skenario Nyata User
// Pastikan server 'simdokpol' sudah jalan sebelum test ini dieksekusi!
func TestEndToEndFlow(t *testing.T) {
	// 1. Tunggu Server Ready (Health Check Manual)
	waitForServer(t)

	// 2. Skenario: Login sebagai Super Admin (Default Seeding)
	token := performLogin(t, "12345678", "admin123")
	fmt.Println("✅ Login Sukses! Token didapat.")

	// 3. Skenario: Cek Dashboard Stats (Butuh Auth)
	performGetDashboard(t, token)
	fmt.Println("✅ Akses Dashboard Sukses!")

	// 4. Skenario: Buat Surat Baru (Create Document)
	performCreateDocument(t, token)
	fmt.Println("✅ Buat Surat Sukses!")
}

func waitForServer(t *testing.T) {
	// Coba ping server sampai 30 detik
	for i := 0; i < 30; i++ {
		resp, err := http.Get(baseURL + "/login")
		if err == nil && resp.StatusCode == 200 {
			return // Server ready
		}
		time.Sleep(1 * time.Second)
	}
	t.Fatal("Server tidak menyala dalam 30 detik!")
}

func performLogin(t *testing.T, nrp, password string) string {
	payload := LoginReq{NRP: nrp, Password: password}
	jsonData, _ := json.Marshal(payload)

	resp, err := http.Post(baseURL+"/api/login", "application/json", bytes.NewBuffer(jsonData))
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode, "Login harus return 200 OK")
	
	// Ambil cookie token
	cookies := resp.Cookies()
	var token string
	for _, cookie := range cookies {
		if cookie.Name == "token" {
			token = cookie.Value
		}
	}
	assert.NotEmpty(t, token, "Token cookie harus ada")
	return token
}

func performGetDashboard(t *testing.T, token string) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", baseURL+"/api/stats", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token})

	resp, err := client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode, "Harus bisa akses dashboard stats")
}

func performCreateDocument(t *testing.T, token string) {
	// Data Dummy Surat
	payload := map[string]interface{}{
		"nama_lengkap":       "WARGA TEST E2E",
		"tempat_lahir":       "JAKARTA",
		"tanggal_lahir":      "01-01-1990",
		"jenis_kelamin":      "Laki-laki",
		"agama":              "Islam",
		"pekerjaan":          "Wiraswasta",
		"alamat":             "Jl. Testing No. 1",
		"lokasi_hilang":      "Pasar Senen",
		"petugas_pelapor_id": 1, // Asumsi ID Admin
		"pejabat_persetuju_id": 1,
		"items": []map[string]string{
			{"nama_barang": "KTP", "deskripsi": "NIK: 3171234567890001"},
		},
	}
	jsonData, _ := json.Marshal(payload)

	client := &http.Client{}
	req, _ := http.NewRequest("POST", baseURL+"/api/documents", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "token", Value: token})

	resp, err := client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Gagal buat surat. Status: %d, Body: %s", resp.StatusCode, string(body))
	}
}