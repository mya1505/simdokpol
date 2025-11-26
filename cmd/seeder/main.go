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

// --- DATA DUMMY ---
var (
	firstNames = []string{"Agus", "Budi", "Citra", "Dewi", "Eko", "Fajar", "Gita", "Hendra", "Indah", "Joko", "Kartika", "Lukman", "Maya", "Nur", "Oki", "Putri", "Rudi", "Siti", "Tono", "Wawan", "Dedi", "Yudi", "Rina", "Sari", "Bambang"}
	lastNames  = []string{"Santoso", "Purnomo", "Wijaya", "Saputra", "Hidayat", "Suryana", "Kusuma", "Pratama", "Setiawan", "Wulandari", "Permana", "Kurniawan", "Nugroho", "Susanti", "Rahayu", "Siregar", "Nasution", "Chaniago", "Wibowo", "Utami"}
	pangkats   = []string{"BRIPDA", "BRIPTU", "BRIGPOL", "BRIPKA", "AIPDA", "AIPTU", "IPDA", "IPTU", "AKP"}
	jabatans   = []string{"ANGGOTA JAGA REGU", "KANIT SPKT", "KA SPKT", "BANIT SPKT", "BA SPKT"}
	regus      = []string{"I", "II", "III"}
	pekerjaans = []string{"Wiraswasta", "Petani/Pekebun", "Nelayan", "Karyawan Swasta", "Pegawai Negeri Sipil", "Pelajar/Mahasiswa", "Buruh Harian Lepas", "Mengurus Rumah Tangga", "Pedagang"}
	locations  = []string{"Pasar Bahodopi", "Jalan Trans Sulawesi", "Pantai Kurisa", "Depan Bank BRI", "Area Parkir PT IMIP", "Warung Makan Jawa", "Masjid Raya", "Lapangan Bola", "Dusun I", "Dusun II", "Dusun III"}

	itemTypes = []struct {
		Name         string
		DescTemplate string
	}{
		{"KTP", "NIK: %s a.n. Pelapor"},
		{"SIM C", "No. SIM: %s a.n. Pelapor"},
		{"SIM A", "No. SIM: %s a.n. Pelapor"},
		{"STNK", "Sepeda Motor Honda Beat, No. Pol: %s"},
		{"Kartu ATM", "Bank BRI Unit Bahodopi, No. Rek: %s"},
		{"BPKB", "Mobil Toyota Avanza, No. BPKB: %s"},
		{"Ijazah SMA", "No. Ijazah: %s, Tahun Lulus 2015"},
		{"Kartu Keluarga", "No. KK: %s a.n. Kepala Keluarga Pelapor"},
	}
)

func main() {
	rand.Seed(time.Now().UnixNano())

	// 1. Load Config
	envPath := filepath.Join(utils.GetAppDataDir(), ".env")
	_ = godotenv.Load(envPath)
	_ = godotenv.Load() // Fallback

	cfg := config.LoadConfig()

	// --- HEADER ---
	fmt.Println("\n========================================")
	fmt.Println("   SIMDOKPOL SMART SEEDER v2.0 🧠    ")
	fmt.Println("========================================")
	fmt.Printf("🔌 Database : %s\n", cfg.DBDialect)
	
	// Input Interaktif
	var userCountInput, docCountInput int
	fmt.Print("\n👉 Jumlah User Tambahan (selain Admin): ")
	fmt.Scanln(&userCountInput)
	if userCountInput <= 0 { userCountInput = 5 }

	fmt.Print("👉 Jumlah Dokumen Dummy: ")
	fmt.Scanln(&docCountInput)
	if docCountInput <= 0 { docCountInput = 50 }
	fmt.Println("\n⏳ Memproses data...")

	// 2. Setup DB
	db := setupDatabase(cfg)
	
	// 3. Seeding
	seedConfigs(db)
	
	// Generate User (Return list user, jumlah yg dibuat, dan satu contoh admin)
	users, usersCreated, demoAdmin := seedUsers(db, cfg.BcryptCost, userCountInput)
	
	// Generate Docs
	docsCreated := seedDocuments(db, users, docCountInput)

	// --- LAPORAN AKHIR ---
	fmt.Println("\n✅ SEEDING SELESAI DENGAN SUKSES!")
	fmt.Println("========================================")
	fmt.Println("📊 RINGKASAN STATISTIK:")
	fmt.Printf("   👤 Total User Dibuat    : %d\n", usersCreated)
	fmt.Printf("   📄 Total Dokumen Dibuat : %d\n", docsCreated)
	fmt.Println("----------------------------------------")
	fmt.Println("🔑 KREDENSIAL LOGIN (SUPER ADMIN):")
	fmt.Printf("   User / NRP : %s\n", demoAdmin.NRP)
	fmt.Printf("   Password   : password\n") // Password hardcoded di seeder
	fmt.Println("========================================")
	fmt.Println("👉 Silakan login menggunakan kredensial di atas.")
}

func setupDatabase(cfg *config.Config) *gorm.DB {
	var db *gorm.DB
	var err error
	
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error), // Silent
	}

	switch cfg.DBDialect {
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName)
		db, err = gorm.Open(mysql.Open(dsn), gormConfig)
	case "postgres":
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
			cfg.DBHost, cfg.DBUser, cfg.DBPass, cfg.DBName, cfg.DBPort)
		db, err = gorm.Open(postgres.Open(dsn), gormConfig)
	default: // sqlite
		db, err = gorm.Open(sqlite.Open(cfg.DBDSN), gormConfig)
		if err == nil { db.Exec("PRAGMA foreign_keys = ON") }
	}

	if err != nil {
		log.Fatalf("❌ Gagal koneksi DB: %v", err)
	}
	
	// Pastikan tabel ada
	db.AutoMigrate(
		&models.User{}, &models.Resident{}, &models.LostDocument{}, 
		&models.LostItem{}, &models.AuditLog{}, &models.Configuration{}, 
		&models.ItemTemplate{}, &models.License{},
	)
	
	return db
}

func seedConfigs(db *gorm.DB) {
	configs := []models.Configuration{
		{Key: "is_setup_complete", Value: "true"},
		{Key: "kop_baris_1", Value: "KEPOLISIAN NEGARA REPUBLIK INDONESIA"},
		{Key: "kop_baris_2", Value: "DAERAH SULAWESI TENGAH"},
		{Key: "kop_baris_3", Value: "RESOR MOROWALI"},
		{Key: "nama_kantor", Value: "POLSEK BAHODOPI"},
		{Key: "tempat_surat", Value: "Bahodopi"},
		{Key: "format_nomor_surat", Value: "SKH/%03d/%s/TUK.7.2.1/%d"},
		{Key: "nomor_surat_terakhir", Value: "0"},
		{Key: "zona_waktu", Value: "Asia/Makassar"},
		{Key: "archive_duration_days", Value: "15"},
		{Key: "license_status", Value: "VALID"},
	}
	for _, c := range configs {
		db.FirstOrCreate(&c, models.Configuration{Key: c.Key})
	}
}

// Mengembalikan: Slice User, Jumlah User Baru, Data Admin untuk Display
func seedUsers(db *gorm.DB, cost int, count int) ([]models.User, int, *models.User) {
	passBytes, _ := bcrypt.GenerateFromPassword([]byte("password"), cost)
	passwordHash := string(passBytes)
	usersCreated := 0

	// 1. Super Admin Fixed (Safety Net)
	admin := models.User{
		NamaLengkap: "ADMINISTRATOR UTAMA",
		NRP:         "12345678", 
		KataSandi:   passwordHash,
		Pangkat:     "AIPDA",
		Peran:       models.RoleSuperAdmin,
		Jabatan:     "KANIT SPKT",
		Regu:        "-",
	}
	
	// Cek dulu biar created count akurat
	var checkAdmin models.User
	if err := db.Where("nrp = ?", admin.NRP).First(&checkAdmin).Error; err != nil {
		db.Create(&admin)
		usersCreated++
	} else {
		admin = checkAdmin // Pakai yg existing
	}

	var allUsers []models.User
	allUsers = append(allUsers, admin)

	// 2. Generate Random Users (Mix Operator & Admin)
	for i := 0; i < count; i++ {
		nrp := fmt.Sprintf("%d", 80000000+rand.Intn(10000000))
		
		// Random Logic: 20% kemungkinan jadi Super Admin tambahan
		peran := models.RoleOperator
		jabatan := jabatans[rand.Intn(len(jabatans))]
		
		if rand.Intn(100) < 20 { // 20% chance
			peran = models.RoleSuperAdmin
			jabatan = "KANIT PROVOS" // Contoh jabatan lain admin
		}

		user := models.User{
			NamaLengkap: randomName(),
			NRP:         nrp,
			KataSandi:   passwordHash,
			Pangkat:     pangkats[rand.Intn(len(pangkats))],
			Peran:       peran,
			Jabatan:     jabatan,
			Regu:        regus[rand.Intn(len(regus))],
		}

		var exist models.User
		if err := db.Where("nrp = ?", nrp).First(&exist).Error; err != nil {
			db.Create(&user)
			allUsers = append(allUsers, user)
			usersCreated++
		} else {
			allUsers = append(allUsers, exist)
		}
	}
	
	return allUsers, usersCreated, &admin
}

func seedDocuments(db *gorm.DB, users []models.User, count int) int {
	var existingCount int64
	db.Model(&models.LostDocument{}).Count(&existingCount)
	startNum := int(existingCount) + 1
	docsCreated := 0

	for i := 0; i < count; i++ {
		daysAgo := rand.Intn(90)
		date := time.Now().AddDate(0, 0, -daysAgo)
		
		operator := users[rand.Intn(len(users))]
		petugas := users[rand.Intn(len(users))]
		
		job := pekerjaans[rand.Intn(len(pekerjaans))]
		nik := fmt.Sprintf("72%02d%02d%02d%02d%02d%04d", 
			rand.Intn(90)+10, rand.Intn(90)+10, rand.Intn(28)+1, 
			rand.Intn(12)+1, rand.Intn(99), rand.Intn(9999))
		
		res := models.Resident{
			NIK:          nik,
			NamaLengkap:  randomName(),
			TempatLahir:  "Morowali",
			TanggalLahir: time.Now().AddDate(-20-rand.Intn(30), 0, 0),
			JenisKelamin: []string{"Laki-laki", "Perempuan"}[rand.Intn(2)],
			Agama:        "Islam",
			Pekerjaan:    job,
			Alamat:       fmt.Sprintf("Desa Bahodopi Dusun %d", rand.Intn(5)+1),
		}
		db.Create(&res)

		status := models.StatusDiterbitkan
		if daysAgo > 15 {
			status = models.StatusDiarsipkan
		}

		docNum := fmt.Sprintf("SKH/%03d/%s/TUK.7.2.1/%d", startNum+i, intToRoman(int(date.Month())), date.Year())
		loc := locations[rand.Intn(len(locations))]

		doc := models.LostDocument{
			NomorSurat:         docNum,
			TanggalLaporan:     date,
			Status:             status,
			LokasiHilang:       loc,
			ResidentID:         res.ID,
			PetugasPelaporID:   petugas.ID,
			PejabatPersetujuID: &users[0].ID, // Ttd oleh Admin utama biar valid
			OperatorID:         operator.ID,
			TanggalPersetujuan: &date,
		}
		
		if err := db.Create(&doc).Error; err == nil {
			docsCreated++
		} else {
			continue
		}

		itemCount := rand.Intn(2) + 1
		for k := 0; k < itemCount; k++ {
			tmpl := itemTypes[rand.Intn(len(itemTypes))]
			randomSerial := strconv.Itoa(10000000 + rand.Intn(90000000))
			
			item := models.LostItem{
				LostDocumentID: doc.ID,
				NamaBarang:     tmpl.Name,
				Deskripsi:      fmt.Sprintf(tmpl.DescTemplate, randomSerial),
			}
			db.Create(&item)
		}

		audit := models.AuditLog{
			UserID:    operator.ID,
			Aksi:      models.AuditCreateDocument,
			Detail:    fmt.Sprintf("Membuat surat keterangan hilang baru dengan nomor: %s", docNum),
			Timestamp: date,
		}
		db.Create(&audit)
	}
	return docsCreated
}

func randomName() string {
	return fmt.Sprintf("%s %s", firstNames[rand.Intn(len(firstNames))], lastNames[rand.Intn(len(lastNames))])
}

func intToRoman(num int) string {
	roman := map[int]string{1: "I", 2: "II", 3: "III", 4: "IV", 5: "V", 6: "VI", 7: "VII", 8: "VIII", 9: "IX", 10: "X", 11: "XI", 12: "XII"}
	return roman[num]
}