package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

func TestEndToEndFlow(t *testing.T) {
	// 1. Tunggu Server Ready (Health Check)
	fmt.Println("⏳ Menunggu server up...")
	waitForServer(t)

	// 2. STEP BARU: Lakukan Setup Awal (Bikin Admin)
	// Karena database kosong, kita harus register admin dulu lewat API Setup
	fmt.Println("🛠️ Melakukan Setup Awal...")
	performSetup(t)
	
	// Tunggu sebentar karena SaveSetup memicu RESTART server
	fmt.Println("🔄 Menunggu server restart setelah setup...")
	time.Sleep(5 * time.Second) 
	waitForServer(t) // Pastikan server up lagi

	// 3. Skenario: Login sebagai Super Admin (Yang barusan dibuat)
	fmt.Println("🔑 Mencoba Login...")
	token := performLogin(t, "12345678", "admin123")
	fmt.Println("✅ Login Sukses! Token didapat.")

	// 4. Skenario: Cek Dashboard Stats
	performGetDashboard(t, token)
	fmt.Println("✅ Akses Dashboard Sukses!")

	// 5. Skenario: Buat Surat Baru
	performCreateDocument(t, token)
	fmt.Println("✅ Buat Surat Sukses!")
}

func waitForServer(t *testing.T) {
	for i := 0; i < 30; i++ {
		// Cek endpoint setup karena ini yang pasti terbuka saat awal
		resp, err := http.Get(baseURL + "/setup")
		if err == nil && resp.StatusCode < 500 {
			return 
		}
		time.Sleep(1 * time.Second)
	}
	t.Fatal("Server tidak menyala dalam 30 detik!")
}

// Fungsi baru untuk nembak API Setup
func performSetup(t *testing.T) {
	// Payload sesuai SaveSetupRequest di ConfigController
	payload := map[string]string{
		"db_dialect": "sqlite", 
		"db_dsn": "e2e_test.db",
		"kop_baris_1": "KEPOLISIAN NEGARA",
		"kop_baris_2": "REPUBLIK INDONESIA",
		"kop_baris_3": "SEKTOR E2E TEST",
		"nama_kantor": "POLSEK E2E",
		"tempat_surat": "JAKARTA",
		"format_nomor_surat": "SKH/%03d/X/2025",
		"nomor_surat_terakhir": "0",
		"zona_waktu": "Asia/Jakarta",
		"archive_duration_days": "30",
		
		// Kredensial Admin yang akan kita pakai login nanti
		"admin_nama_lengkap": "Super Admin E2E",
		"admin_nrp": "12345678",
		"admin_pangkat": "JENDERAL",
		"admin_password": "admin123",
	}
	
	jsonData, _ := json.Marshal(payload)
	resp, err := http.Post(baseURL+"/api/setup", "application/json", bytes.NewBuffer(jsonData))
	assert.NoError(t, err)
	defer resp.Body.Close()

	// Setup mungkin mengembalikan 200 OK
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		// Abaikan error jika setup sudah done (403), lanjut ke login
		if resp.StatusCode == 403 {
			fmt.Println("⚠️ Setup sudah dilakukan sebelumnya, lanjut login.")
			return
		}
		t.Fatalf("Gagal Setup. Status: %d, Body: %s", resp.StatusCode, string(body))
	}
}

func performLogin(t *testing.T, nrp, password string) string {
	payload := LoginReq{NRP: nrp, Password: password}
	jsonData, _ := json.Marshal(payload)

	resp, err := http.Post(baseURL+"/api/login", "application/json", bytes.NewBuffer(jsonData))
	assert.NoError(t, err)
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Gagal Login. Status: %d, Body: %s", resp.StatusCode, string(body))
	}

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

	assert.Equal(t, 200, resp.StatusCode)
}

func performCreateDocument(t *testing.T, token string) {
	payload := map[string]interface{}{
		"nama_lengkap":       "WARGA TEST E2E",
		"tempat_lahir":       "JAKARTA",
		"tanggal_lahir":      "01-01-1990",
		"jenis_kelamin":      "Laki-laki",
		"agama":              "Islam",
		"pekerjaan":          "Wiraswasta",
		"alamat":             "Jl. Testing No. 1",
		"lokasi_hilang":      "Pasar Senen",
		"petugas_pelapor_id": 1, // ID Admin yang baru dibuat pasti 1
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