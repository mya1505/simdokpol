package main

import (
	"fmt"
	"log"
	"math/rand"
	"path/filepath"
	"simdokpol/internal/config"
	"simdokpol/internal/models"
	"simdokpol/internal/utils"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ... (Variable data dummy sama persis) ...
var (
	firstNames = []string{"Agus", "Budi", "Citra", "Dewi", "Eko", "Fajar", "Gita", "Hendra", "Indah", "Joko", "Kartika", "Lukman", "Maya", "Nur", "Oki", "Putri", "Rudi", "Siti", "Tono", "Wawan", "Dedi", "Yudi", "Rina", "Sari", "Bambang"}
	lastNames  = []string{"Santoso", "Purnomo", "Wijaya", "Saputra", "Hidayat", "Suryana", "Kusuma", "Pratama", "Setiawan", "Wulandari", "Permana", "Kurniawan", "Nugroho", "Susanti", "Rahayu", "Siregar", "Nasution", "Chaniago", "Wibowo", "Utami"}
	pangkats = []string{"BRIPDA", "BRIPTU", "BRIGPOL", "BRIPKA", "AIPDA", "AIPTU", "IPDA", "IPTU", "AKP"}
	jabatans = []string{"ANGGOTA JAGA REGU", "KANIT SPKT", "KA SPKT", "BANIT SPKT", "BA SPKT"}
	regus    = []string{"I", "II", "III"}
	pekerjaans = []string{"Wiraswasta", "Petani/Pekebun", "Nelayan", "Karyawan Swasta", "Pegawai Negeri Sipil", "Pelajar/Mahasiswa", "Buruh Harian Lepas", "Mengurus Rumah Tangga", "Pedagang"}
	locations  = []string{"Pasar Bahodopi", "Jalan Trans Sulawesi", "Pantai Kurisa", "Depan Bank BRI", "Area Parkir PT IMIP", "Warung Makan Jawa", "Masjid Raya", "Lapangan Bola", "Dusun I", "Dusun II", "Dusun III"}
	itemTypes = []struct {
		Name string
		DescTemplate string
	}{
		{"KTP", "NIK: %s a.n. Pelapor"},
		{"SIM C", "No. SIM: %s a.n. Pelapor"},
		{"STNK", "Sepeda Motor Honda Beat, No. Pol: %s"},
		{"BPKB", "Mobil Toyota Avanza, No. BPKB: %s"},
	}
)

func main() {
	rand.Seed(time.Now().UnixNano())

	// --- FIX: Load Env dari AppData (Sama dengan Main App) ---
	envPath := filepath.Join(utils.GetAppDataDir(), ".env")
	_ = godotenv.Load(envPath)

	// --- FIX: Gunakan Config Loader yang sudah diperbaiki ---
	cfg := config.LoadConfig()

	fmt.Println("=== SIMDOKPOL SEEDER ===")
	fmt.Printf("Target DB: %s\n", cfg.DBDialect)
	// Pastikan DSN mengarah ke file di AppData
	if cfg.DBDialect == "sqlite" {
		fmt.Printf("File Path: %s\n", cfg.DBDSN)
	}

	var userCountInput, docCountInput int
	fmt.Print("Jumlah User Tambahan (Default 5): ")
	fmt.Scanln(&userCountInput)
	if userCountInput <= 0 { userCountInput = 5 }

	fmt.Print("Jumlah Dokumen (Default 50): ")
	fmt.Scanln(&docCountInput)
	if docCountInput <= 0 { docCountInput = 50 }

	db := setupDatabase(cfg)
	
	seedConfigs(db)
	users := seedUsers(db, cfg.BcryptCost, userCountInput)
	seedDocuments(db, users, docCountInput)

	fmt.Println("✅ SELESAI!")
}

// ... (Fungsi setupDatabase, seedConfigs, seedUsers SAMA PERSIS dengan kode sebelumnya) ...
// ... (Hanya pastikan import 'config' mengacu ke internal/config yang baru) ...

func setupDatabase(cfg *config.Config) *gorm.DB {
	var db *gorm.DB
	var err error
	gormConfig := &gorm.Config{ Logger: logger.Default.LogMode(logger.Error) }

	switch cfg.DBDialect {
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName)
		db, err = gorm.Open(mysql.Open(dsn), gormConfig)
	case "postgres":
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Jakarta", cfg.DBHost, cfg.DBUser, cfg.DBPass, cfg.DBName, cfg.DBPort, cfg.DBSSLMode)
		db, err = gorm.Open(postgres.Open(dsn), gormConfig)
	default: 
		// SQLite DSN sudah fix dari config.LoadConfig()
		db, err = gorm.Open(sqlite.Open(cfg.DBDSN), gormConfig)
		if err == nil { db.Exec("PRAGMA foreign_keys = ON") }
	}

	if err != nil { log.Fatal("Gagal DB:", err) }
	db.AutoMigrate(&models.User{}, &models.Resident{}, &models.LostDocument{}, &models.LostItem{}, &models.AuditLog{}, &models.Configuration{}, &models.ItemTemplate{}, &models.License{})
	return db
}

// ... (Fungsi seedConfigs, seedUsers, seedDocuments COPY PASTE AJA YANG LAMA, LOGICNYA UDAH BENER) ...
// ... (Yang penting di atas: LoadConfig dan setupDatabase pakai cfg yang benar) ...

func seedConfigs(db *gorm.DB) {
	configs := []models.Configuration{
		{Key: "is_setup_complete", Value: "true"},
		{Key: "kop_baris_1", Value: "KEPOLISIAN NEGARA REPUBLIK INDONESIA"},
		{Key: "nama_kantor", Value: "POLSEK BAHODOPI"},
		{Key: "format_nomor_surat", Value: "SKH/%03d/%s/TUK.7.2.1/%d"},
		{Key: "nomor_surat_terakhir", Value: "0"},
		{Key: "license_status", Value: "VALID"},
	}
	for _, c := range configs { db.FirstOrCreate(&c, models.Configuration{Key: c.Key}) }
}

func seedUsers(db *gorm.DB, cost int, count int) []models.User {
	pass, _ := bcrypt.GenerateFromPassword([]byte("password"), cost)
	admin := models.User{NamaLengkap: "ADMINISTRATOR", NRP: "12345678", KataSandi: string(pass), Peran: models.RoleSuperAdmin, Jabatan: "KANIT SPKT"}
	db.Where(models.User{NRP: admin.NRP}).FirstOrCreate(&admin)
	
	var users []models.User
	users = append(users, admin)
	
	for i:=0; i<count; i++ {
		nrp := fmt.Sprintf("%d", 80000000+rand.Intn(10000000))
		u := models.User{NamaLengkap: randomName(), NRP: nrp, KataSandi: string(pass), Peran: models.RoleOperator, Jabatan: "ANGGOTA JAGA REGU", Regu: "I", Pangkat: "BRIPDA"}
		if err := db.Create(&u).Error; err == nil { users = append(users, u) }
	}
	return users
}

func seedDocuments(db *gorm.DB, users []models.User, count int) {
	for i:=0; i<count; i++ {
		date := time.Now().AddDate(0, 0, -rand.Intn(30))
		res := models.Resident{NIK: fmt.Sprintf("720%d", rand.Int63()), NamaLengkap: randomName(), Alamat: "Bahodopi"}
		db.Create(&res)
		
		doc := models.LostDocument{
			NomorSurat: fmt.Sprintf("SKH/DUMMY/%d", i), TanggalLaporan: date, Status: "DITERBITKAN",
			ResidentID: res.ID, OperatorID: users[0].ID, PetugasPelaporID: users[0].ID,
		}
		db.Create(&doc)
		
		db.Create(&models.AuditLog{UserID: users[0].ID, Aksi: "BUAT DOKUMEN", Timestamp: date})
	}
}

func randomName() string { return firstNames[rand.Intn(len(firstNames))] + " " + lastNames[rand.Intn(len(lastNames))] }