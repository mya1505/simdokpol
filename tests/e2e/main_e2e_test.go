package e2e

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time" // <-- SUDAH DIPASTIKAN ADA

	"github.com/stretchr/testify/assert"
)

const (
	baseURL       = "http://localhost:8080"
	// Secret key default untuk dev/test environment
	testSecretKey = "JANGAN_PAKAI_DEFAULT_KEY_INI_BAHAYA" 
)

type LoginReq struct {
	NRP      string `json:"nrp"`
	Password string `json:"password"`
}

func TestEndToEndFlow(t *testing.T) {
	fmt.Println("⏳ [1/10] Menunggu server up...")
	waitForServer(t)

	fmt.Println("🛠️ [2/10] Setup Awal (Super Admin)...")
	performSetup(t)
	
	fmt.Println("🔄 Menunggu server restart...")
	time.Sleep(5 * time.Second) 
	waitForServer(t)

	fmt.Println("🔑 [3/10] Login Admin...")
	token := performLogin(t, "12345678", "admin123")
	
	// --- TEST SUITE LENGKAP ---

	fmt.Println("🔐 [4/10] Testing Aktivasi Lisensi PRO...")
	performLicenseActivation(t, token)

	fmt.Println("📊 [5/10] Testing Dashboard & Stats...")
	performGetDashboard(t, token)

	fmt.Println("👥 [6/10] Testing Manajemen User (Create Operator)...")
	performUserManagement(t, token)

	fmt.Println("🧩 [7/10] Testing Template Barang (Fitur Pro)...")
	performTemplateManagement(t, token)

	fmt.Println("🌐 [8/10] Testing Konfigurasi HTTPS...")
	performHTTPSCheck(t, token)

	fmt.Println("📝 [9/10] Testing Dokumen (Create SKTLK)...")
	performCreateDocument(t, token)

	fmt.Println("⚙️ [10/10] Testing Utilities (Report/Backup)...")
	performSettingsAndUtils(t, token)

	fmt.Println("✅✅✅ ULTIMATE TEST SUCCESS! SYSTEM 100% STABIL. ✅✅✅")
}

// --- HELPER FUNCTIONS ---

func waitForServer(t *testing.T) {
	for i := 0; i < 30; i++ {
		resp, err := http.Get(baseURL + "/setup")
		if err == nil && resp.StatusCode < 500 {
			return 
		}
		time.Sleep(1 * time.Second)
	}
	t.Fatal("Server tidak menyala dalam 30 detik!")
}

func performSetup(t *testing.T) {
	// FIX: Gunakan payload DSN kosong agar backend menggunakan default path
	payload := map[string]string{
		"db_dialect": "sqlite", 
		"db_dsn": "", 
		"kop_baris_1": "KEPOLISIAN NEGARA",
		"kop_baris_2": "REPUBLIK INDONESIA",
		"kop_baris_3": "RESOR TESTING",
		"nama_kantor": "POLSEK E2E ULTIMATE",
		"tempat_surat": "JAKARTA",
		"format_nomor_surat": "SKH/%03d/TEST/2025",
		"nomor_surat_terakhir": "0",
		"zona_waktu": "Asia/Jakarta",
		"archive_duration_days": "30",
		"admin_nama_lengkap": "Super Admin E2E",
		"admin_nrp": "12345678",
		"admin_pangkat": "JENDERAL",
		"admin_password": "admin123",
	}
	
	jsonData, _ := json.Marshal(payload)
	resp, err := http.Post(baseURL+"/api/setup", "application/json", bytes.NewBuffer(jsonData))
	assert.NoError(t, err)
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode == 403 { return } 
		t.Fatalf("Gagal Setup: %s", string(body))
	}
}

func performLogin(t *testing.T, nrp, password string) string {
	payload := LoginReq{NRP: nrp, Password: password}
	jsonData, _ := json.Marshal(payload)

	resp, err := http.Post(baseURL+"/api/login", "application/json", bytes.NewBuffer(jsonData))
	assert.NoError(t, err)
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("Gagal Login Status: %d", resp.StatusCode)
	}

	cookies := resp.Cookies()
	for _, cookie := range cookies {
		if cookie.Name == "token" {
			return cookie.Value
		}
	}
	t.Fatal("Token cookie tidak ditemukan")
	return ""
}

func performLicenseActivation(t *testing.T, token string) {
	client := &http.Client{}
	
	reqHWID, _ := http.NewRequest("GET", baseURL+"/api/license/hwid", nil)
	reqHWID.AddCookie(&http.Cookie{Name: "token", Value: token})
	respHWID, err := client.Do(reqHWID)
	assert.NoError(t, err)
	
	var hwidResp map[string]string
	json.NewDecoder(respHWID.Body).Decode(&hwidResp)
	hwid := hwidResp["hardware_id"]
	assert.NotEmpty(t, hwid, "HWID tidak boleh kosong")

	h := hmac.New(sha256.New, []byte(testSecretKey))
	h.Write([]byte(hwid))
	hash := h.Sum(nil)
	rawKey := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(hash[:15])
	var formattedKey strings.Builder
	for i, r := range rawKey {
		if i > 0 && i%5 == 0 { formattedKey.WriteRune('-') }
		formattedKey.WriteRune(r)
	}
	validKey := formattedKey.String()

	payload := map[string]string{"key": validKey}
	jsonPayload, _ := json.Marshal(payload)
	reqAct, _ := http.NewRequest("POST", baseURL+"/api/license/activate", bytes.NewBuffer(jsonPayload))
	reqAct.Header.Set("Content-Type", "application/json")
	reqAct.AddCookie(&http.Cookie{Name: "token", Value: token})

	respAct, err := client.Do(reqAct)
	assert.NoError(t, err)
	
	if respAct.StatusCode != 200 {
		body, _ := io.ReadAll(respAct.Body)
		t.Fatalf("Gagal Aktivasi Lisensi: %s", string(body))
	}
	fmt.Println("   ✅ Lisensi Validated & Activated")
}

func performHTTPSCheck(t *testing.T, token string) {
	payload := map[string]interface{}{
		"enable_https": "true",
		"nama_kantor": "POLSEK E2E SECURE", 
	}
	jsonData, _ := json.Marshal(payload)

	req, _ := http.NewRequest("PUT", baseURL+"/api/settings", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "token", Value: token})

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)
	
	var response map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&response)

	assert.Equal(t, true, response["check_https_cert"])
	assert.Equal(t, true, response["restart_required"])
	
	fmt.Println("   ✅ HTTPS Config Logic Verified")
}

func performGetDashboard(t *testing.T, token string) {
	req, _ := http.NewRequest("GET", baseURL+"/api/stats", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token})
	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func performUserManagement(t *testing.T, token string) {
	payload := map[string]string{
		"nama_lengkap": "Petugas E2E",
		"nrp": "88888888",
		"kata_sandi": "petugas123",
		"pangkat": "BRIPDA",
		"peran": "OPERATOR",
		"jabatan": "ANGGOTA JAGA",
		"regu": "I",
	}
	jsonData, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", baseURL+"/api/users", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "token", Value: token})

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode)
}

func performTemplateManagement(t *testing.T, token string) {
	payload := map[string]interface{}{
		"nama_barang": "LAPTOP GAMING",
		"urutan": 10,
		"status": "Aktif",
		"fields_config": []map[string]interface{}{
			{"label": "Merk", "type": "text", "data_label": "Merk"},
		},
	}
	jsonData, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", baseURL+"/api/item-templates", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "token", Value: token})

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode)
}

func performCreateDocument(t *testing.T, token string) {
	payload := map[string]interface{}{
		"nama_lengkap": "WARGA TEST E2E",
		"tempat_lahir": "JAKARTA",
		"tanggal_lahir": "1990-01-01", // Format YYYY-MM-DD
		"jenis_kelamin": "Laki-laki",
		"agama": "Islam",
		"pekerjaan": "Wiraswasta",
		"alamat": "Jl. Testing No. 1",
		"lokasi_hilang": "Pasar Senen",
		"petugas_pelapor_id": 1,
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
	if resp.StatusCode != 201 {
		t.Fatalf("Gagal Buat Surat: %d", resp.StatusCode)
	}
}

func performSettingsAndUtils(t *testing.T, token string) {
	client := &http.Client{}

	reqBack, _ := http.NewRequest("POST", baseURL+"/api/backups", nil)
	reqBack.AddCookie(&http.Cookie{Name: "token", Value: token})
	respBack, err := client.Do(reqBack)
	assert.NoError(t, err)
	assert.Equal(t, 200, respBack.StatusCode)

	startDate := time.Now().Format("2006-01-02")
	endDate := time.Now().Format("2006-01-02")
	reportUrl := fmt.Sprintf("%s/api/reports/aggregate/pdf?start_date=%s&end_date=%s", baseURL, startDate, endDate)
	
	reqRep, _ := http.NewRequest("GET", reportUrl, nil)
	reqRep.AddCookie(&http.Cookie{Name: "token", Value: token})
	respRep, err := client.Do(reqRep)
	assert.NoError(t, err)
	
	if respRep.StatusCode == 200 {
		assert.Equal(t, "application/pdf", respRep.Header.Get("Content-Type"))
		fmt.Println("   ✅ Report PDF Generated (License Active)")
	} else {
		t.Errorf("Gagal Report PDF: Status %d", respRep.StatusCode)
	}
}