package main

import (
	"fmt"
	"log"
	"math/rand"
	"path/filepath"
	"simdokpol/internal/config"
	"simdokpol/internal/models"
	"simdokpol/internal/utils"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ============================================================================
// DATA DUMMY
// ============================================================================

var (
	firstNames = []string{"Agus", "Budi", "Citra", "Dewi", "Eko", "Fajar", "Gita", "Hendra", "Indah", "Joko", "Kartika", "Lukman", "Maya", "Nur", "Oki", "Putri", "Rudi", "Siti", "Tono", "Wawan", "Dedi", "Yudi", "Rina", "Sari", "Bambang", "Slamet", "Widya", "Reza", "Dian", "Faisal"}
	lastNames  = []string{"Santoso", "Purnomo", "Wijaya", "Saputra", "Hidayat", "Suryana", "Kusuma", "Pratama", "Setiawan", "Wulandari", "Permana", "Kurniawan", "Nugroho", "Susanti", "Rahayu", "Siregar", "Nasution", "Chaniago", "Wibowo", "Utami", "Firmansyah", "Ramadhan", "Haryanto"}
	
	pangkatPerwira = []string{"IPDA", "IPTU", "AKP", "KOMPOL"}
	pangkatBintara = []string{"BRIPDA", "BRIPTU", "BRIGPOL", "BRIPKA", "AIPDA", "AIPTU"}
	
	jabatanKanitRegu   = "KANIT JAGA"
	jabatanAnggotaJaga = "ANGGOTA JAGA REGU"
	
	regus = []string{"I", "II", "III"}
	
	agamaList  = []string{"Islam", "Kristen Protestan", "Katolik", "Hindu", "Buddha", "Konghucu"}
	pekerjaans = []string{"Wiraswasta", "Petani/Pekebun", "Nelayan", "Karyawan Swasta", "Pegawai Negeri Sipil", "Pelajar/Mahasiswa", "Buruh Harian Lepas", "Mengurus Rumah Tangga", "Pedagang"}
	locations  = []string{"Pasar Bahodopi", "Jalan Trans Sulawesi", "Pantai Kurisa", "Depan Bank BRI", "Area Parkir PT IMIP", "Warung Makan Jawa", "Masjid Raya", "Lapangan Bola", "Dusun I", "Dusun II", "Dusun III"}

	itemTypes = []struct {
		Name         string
		DescTemplate string
	}{
		{"KTP", "NIK: %s a.n. %s"},
		{"SIM", "No. SIM: %s Golongan %s a.n. %s"},
		{"STNK", "Sepeda Motor %s, No. Pol: %s a.n. %s"},
		{"ATM", "Bank %s, No. Rek: %s a.n. %s"},
		{"BPKB", "Mobil %s, No. BPKB: %s a.n. %s"},
		{"IJAZAH", "Ijazah %s No: %s a.n. %s"},
	}

	simGolongan = []string{"A", "B I", "B II", "C", "D"}
	motorBrands = []string{"Honda Beat", "Yamaha Mio", "Honda Vario", "Suzuki Nex", "Honda Scoopy", "Yamaha Aerox"}
	mobilBrands = []string{"Toyota Avanza", "Daihatsu Xenia", "Suzuki Ertiga", "Honda Mobilio"}
	
	bankNames = []string{
		"BRI", "BCA", "Mandiri", "BNI", "BSI", "BTN", "CIMB Niaga",
		"Danamon", "Permata", "Panin", "Maybank", "Mega",
		"BTPN / Jenius", "Bank Daerah (BPD)",
	}
	
	ijazahLevels = []string{"SD", "SMP", "SMA", "SMK", "D3", "S1", "S2"}
)

// ============================================================================
// MAIN FUNCTION
// ============================================================================

func main() {
	startTime := time.Now()
	rand.Seed(time.Now().UnixNano())

	envPath := filepath.Join(utils.GetAppDataDir(), ".env")
	_ = godotenv.Load(envPath)
	_ = godotenv.Load()

	cfg := config.LoadConfig()

	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("          🚀 SIMDOKPOL SEEDER v2.0 (SYNCHRONIZED)")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Printf("🔌 Database: %s\n", cfg.DBDialect)

	var anggotaPerRegu, docCountInput int
	var withInactive string

	fmt.Print("\n👉 Jumlah Anggota per Regu (min 3): ")
	fmt.Scanln(&anggotaPerRegu)
	if anggotaPerRegu < 3 {
		anggotaPerRegu = 3
	}

	fmt.Print("👉 Jumlah Dokumen Dummy (min 10): ")
	fmt.Scanln(&docCountInput)
	if docCountInput < 10 {
		docCountInput = 50
	}

	fmt.Print("👉 Tambahkan user non-aktif? (y/n): ")
	fmt.Scanln(&withInactive)
	includeInactive := strings.ToLower(withInactive) == "y"

	fmt.Println("\n⏳ Memproses data...")

	db := setupDatabase(cfg)

	seedConfigurations(db)
	seedTemplates(db)
	users, userStats := seedSmartUsers(db, cfg.BcryptCost, anggotaPerRegu, includeInactive)
	docStats := seedEnhancedDocuments(db, users, docCountInput)

	fmt.Println("\n✅ SEEDING BERHASIL!")
	fmt.Println(strings.Repeat("-", 70))
	fmt.Printf("👤 Total User    : %d\n", userStats.Total)
	fmt.Printf("👮 Kanit         : %d\n", userStats.KanitCount)
	fmt.Printf("👮 Anggota       : %d\n", userStats.AnggotaCount)
	fmt.Printf("📄 Dokumen       : %d\n", docStats.Total)
	fmt.Printf("⏱️  Waktu         : %v\n", time.Since(startTime))
	fmt.Println(strings.Repeat("-", 70))
	
	if userStats.SampleAdmin != nil {
		fmt.Println("🔑 SUPER ADMIN LOGIN:")
		fmt.Printf("   User : %s\n", userStats.SampleAdmin.NRP)
		fmt.Printf("   Pass : %s\n", userStats.SampleAdmin.NRP)
	}
	fmt.Println("💡 Semua user lain juga memiliki password = NRP mereka.")
	fmt.Println(strings.Repeat("=", 70))
}

// ============================================================================
// DATABASE & SEEDING LOGIC
// ============================================================================

func setupDatabase(cfg *config.Config) *gorm.DB {
	var db *gorm.DB
	var err error
	gormConfig := &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)}

	switch cfg.DBDialect {
	case "mysql":
		var tlsOption string
		switch cfg.DBSSLMode {
		case "require", "verify-full":
			tlsOption = "true"
		default:
			tlsOption = "false"
		}
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&tls=%s", cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName, tlsOption)
		db, err = gorm.Open(mysql.Open(dsn), gormConfig)
	case "postgres":
		sslMode := cfg.DBSSLMode
		if sslMode == "" {
			sslMode = "disable"
		}
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Jakarta", cfg.DBHost, cfg.DBUser, cfg.DBPass, cfg.DBName, cfg.DBPort, sslMode)
		db, err = gorm.Open(postgres.Open(dsn), gormConfig)
	default:
		db, err = gorm.Open(sqlite.Open(cfg.DBDSN), gormConfig)
		if err == nil {
			db.Exec("PRAGMA journal_mode = WAL;")
			db.Exec("PRAGMA synchronous = NORMAL;")
			db.Exec("PRAGMA foreign_keys = ON;")
		}
	}

	if err != nil {
		log.Fatalf("❌ Gagal koneksi database: %v", err)
	}
	
	db.AutoMigrate(&models.User{}, &models.Resident{}, &models.LostDocument{}, &models.LostItem{}, &models.AuditLog{}, &models.Configuration{}, &models.ItemTemplate{}, &models.License{})
	return db
}

func seedConfigurations(db *gorm.DB) {
	fmt.Println("   ↳ Menyuntikkan Pengaturan Umum...")
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
		db.Where("key = ?", c.Key).FirstOrCreate(&c)
	}
}

func seedTemplates(db *gorm.DB) {
	fmt.Println("   ↳ Menyuntikkan Template Barang (Synchronized)...")
	
	bankOptions := []string{
		"BRI", "BCA", "Mandiri", "BNI", "BSI", "BTN",
		"CIMB Niaga", "Danamon", "Permata", "Panin", "Maybank",
		"Mega", "BTPN / Jenius", "Bank Daerah (BPD)", "Lainnya",
	}

	templates := []models.ItemTemplate{
		{
			NamaBarang: "KTP",
			Urutan:     1,
			IsActive:   true,
			FieldsConfig: models.JSONFieldArray{
				{
					Label:          "NIK",
					Type:           "text",
					DataLabel:      "NIK",
					Regex:          "^[0-9]{16}$",
					RequiredLength: 16,
					IsNumeric:      true,
					Placeholder:    "16 Digit NIK",
				},
			},
		},
		{
			NamaBarang: "SIM",
			Urutan:     2,
			IsActive:   true,
			FieldsConfig: models.JSONFieldArray{
				{
					Label:     "Golongan SIM",
					Type:      "select",
					DataLabel: "Gol",
					Options:   []string{"A", "B I", "B II", "C", "D"},
				},
				{
					Label:     "Nomor SIM",
					Type:      "text",
					DataLabel: "No. SIM",
					Regex:     "^[0-9]{12,14}$",
					MinLength: 12,
					MaxLength: 14,
					IsNumeric: true,
				},
			},
		},
		{
			NamaBarang: "STNK",
			Urutan:     3,
			IsActive:   true,
			FieldsConfig: models.JSONFieldArray{
				{
					Label:     "Nomor Polisi",
					Type:      "text",
					DataLabel: "No. Pol",
				},
				{
					Label:     "Nomor Rangka",
					Type:      "text",
					DataLabel: "No. Rangka",
				},
				{
					Label:     "Nomor Mesin",
					Type:      "text",
					DataLabel: "No. Mesin",
				},
			},
		},
		{
			NamaBarang: "BPKB",
			Urutan:     4,
			IsActive:   true,
			FieldsConfig: models.JSONFieldArray{
				{
					Label:     "Nomor BPKB",
					Type:      "text",
					DataLabel: "No. BPKB",
				},
				{
					Label:     "Atas Nama",
					Type:      "text",
					DataLabel: "a.n.",
				},
			},
		},
		{
			NamaBarang: "IJAZAH",
			Urutan:     5,
			IsActive:   true,
			FieldsConfig: models.JSONFieldArray{
				{
					Label:     "Tingkat",
					Type:      "select",
					DataLabel: "Tingkat",
					Options:   []string{"SD", "SMP", "SMA", "D3", "S1", "S2"},
				},
				{
					Label:     "Nomor Ijazah",
					Type:      "text",
					DataLabel: "No. Ijazah",
				},
			},
		},
		{
			NamaBarang: "ATM",
			Urutan:     6,
			IsActive:   true,
			FieldsConfig: models.JSONFieldArray{
				{
					Label:     "Nama Bank",
					Type:      "select",
					DataLabel: "Bank",
					Options:   bankOptions,
				},
				{
					Label:     "Nomor Rekening",
					Type:      "text",
					DataLabel: "No. Rek",
				},
			},
		},
		{
			NamaBarang:   "LAINNYA",
			Urutan:       99,
			IsActive:     true,
			FieldsConfig: models.JSONFieldArray{},
		},
	}
	
	for _, t := range templates {
		var existing models.ItemTemplate
		if err := db.Where("nama_barang = ?", t.NamaBarang).First(&existing).Error; err == nil {
			existing.FieldsConfig = t.FieldsConfig
			existing.Urutan = t.Urutan
			existing.IsActive = t.IsActive
			db.Save(&existing)
		} else {
			db.Create(&t)
		}
	}
}

type UserSeederStats struct {
	Total, SuperAdmins, Operators, KanitCount, AnggotaCount, Active, Inactive int
	PerRegu     map[string]int
	SampleAdmin *models.User
}

func seedSmartUsers(db *gorm.DB, bcryptCost int, anggotaPerRegu int, includeInactive bool) ([]models.User, *UserSeederStats) {
	stats := &UserSeederStats{PerRegu: make(map[string]int)}
	var allUsers []models.User
	usedNRPs := make(map[string]bool)

	hashPassword := func(nrp string) string {
		bytes, _ := bcrypt.GenerateFromPassword([]byte(nrp), bcryptCost)
		return string(bytes)
	}

	generateUniqueNRP := func() string {
		for {
			nrp := fmt.Sprintf("%d", 80000000+rand.Intn(19999999))
			if !usedNRPs[nrp] {
				usedNRPs[nrp] = true
				return nrp
			}
		}
	}

	sysAdminNRP := "12345678"
	sysAdmin := models.User{
		NamaLengkap: "SYSTEM ADMINISTRATOR",
		NRP:         sysAdminNRP,
		KataSandi:   hashPassword(sysAdminNRP),
		Pangkat:     "-",
		Peran:       models.RoleSuperAdmin,
		Jabatan:     "SUPER ADMIN",
		Regu:        "-",
	}
	db.Where("nrp = ?", sysAdmin.NRP).FirstOrCreate(&sysAdmin)
	db.Model(&sysAdmin).Update("kata_sandi", hashPassword(sysAdminNRP))
	
	stats.Total++
	stats.SuperAdmins++
	stats.Active++
	stats.SampleAdmin = &sysAdmin
	usedNRPs[sysAdminNRP] = true

	for _, regu := range regus {
		kanitNRP := generateUniqueNRP()
		kanit := models.User{
			NamaLengkap: randomName(),
			NRP:         kanitNRP,
			KataSandi:   hashPassword(kanitNRP),
			Pangkat:     pangkatPerwira[rand.Intn(len(pangkatPerwira))],
			Peran:       models.RoleOperator,
			Jabatan:     jabatanKanitRegu,
			Regu:        regu,
		}
		if err := db.Create(&kanit).Error; err == nil {
			allUsers = append(allUsers, kanit)
			stats.Total++
			stats.Operators++
			stats.KanitCount++
			stats.Active++
			stats.PerRegu[regu]++
		}

		for j := 0; j < anggotaPerRegu; j++ {
			isActive := true
			if includeInactive && rand.Float32() < 0.20 {
				isActive = false
			}

			anggotaNRP := generateUniqueNRP()
			anggota := models.User{
				NamaLengkap: randomName(),
				NRP:         anggotaNRP,
				KataSandi:   hashPassword(anggotaNRP),
				Pangkat:     pangkatBintara[rand.Intn(len(pangkatBintara))],
				Peran:       models.RoleOperator,
				Jabatan:     jabatanAnggotaJaga,
				Regu:        regu,
			}
			if !isActive {
				anggota.DeletedAt = gorm.DeletedAt{Time: time.Now(), Valid: true}
			}

			if err := db.Create(&anggota).Error; err == nil {
				allUsers = append(allUsers, anggota)
				stats.Total++
				stats.Operators++
				stats.AnggotaCount++
				stats.PerRegu[regu]++
				if isActive {
					stats.Active++
				} else {
					stats.Inactive++
				}
			}
		}
	}
	return allUsers, stats
}

type DocumentSeederStats struct {
	Total, Published, Archived, Residents, Items, AuditLogs int
}

func seedEnhancedDocuments(db *gorm.DB, users []models.User, count int) *DocumentSeederStats {
	stats := &DocumentSeederStats{}
	var activeOperators, activeKanits []models.User

	for _, u := range users {
		if u.DeletedAt.Valid {
			continue
		}
		if u.Jabatan == jabatanAnggotaJaga {
			activeOperators = append(activeOperators, u)
		}
		if u.Jabatan == jabatanKanitRegu {
			activeKanits = append(activeKanits, u)
		}
	}

	if len(activeOperators) == 0 || len(activeKanits) == 0 {
		return stats
	}

	var existingCount int64
	db.Model(&models.LostDocument{}).Count(&existingCount)
	startNum := int(existingCount) + 1

	for i := 0; i < count; i++ {
		daysAgo := rand.Intn(90)
		docDate := time.Now().AddDate(0, 0, -daysAgo)
		op := activeOperators[rand.Intn(len(activeOperators))]
		
		var pj models.User
		sameRegu := []models.User{}
		for _, k := range activeKanits {
			if k.Regu == op.Regu {
				sameRegu = append(sameRegu, k)
			}
		}
		if len(sameRegu) > 0 {
			pj = sameRegu[rand.Intn(len(sameRegu))]
		} else {
			pj = activeKanits[rand.Intn(len(activeKanits))]
		}

		resident := generateResident()
		if err := db.Create(&resident).Error; err != nil {
			continue
		}
		stats.Residents++

		status := models.StatusDiterbitkan
		if daysAgo > 15 {
			status = models.StatusDiarsipkan
			stats.Archived++
		} else {
			stats.Published++
		}

		docNum := fmt.Sprintf("SKH/%03d/%s/TUK.7.2.1/%d", startNum+i, intToRoman(int(docDate.Month())), docDate.Year())

		doc := models.LostDocument{
			NomorSurat:          docNum,
			TanggalLaporan:      docDate,
			Status:              status,
			LokasiHilang:        locations[rand.Intn(len(locations))],
			ResidentID:          resident.ID,
			PetugasPelaporID:    op.ID,
			PejabatPersetujuID:  &pj.ID,
			OperatorID:          op.ID,
			TanggalPersetujuan:  &docDate,
		}

		if err := db.Create(&doc).Error; err == nil {
			stats.Total++
			item := generateLostItem(doc.ID, resident.NamaLengkap)
			if err := db.Create(&item).Error; err == nil {
				stats.Items++
			}
			
			db.Create(&models.AuditLog{
				UserID:    op.ID,
				Aksi:      models.AuditCreateDocument,
				Detail:    fmt.Sprintf("Membuat surat: %s", docNum),
				Timestamp: docDate,
			})
			stats.AuditLogs++
		}
	}
	return stats
}

func randomName() string {
	return fmt.Sprintf("%s %s", firstNames[rand.Intn(len(firstNames))], lastNames[rand.Intn(len(lastNames))])
}

func generateResident() *models.Resident {
	nik := fmt.Sprintf("72%02d%02d%02d%02d%02d%04d", rand.Intn(90)+10, rand.Intn(90)+10, rand.Intn(28)+1, rand.Intn(12)+1, rand.Intn(99), rand.Intn(9999))
	return &models.Resident{
		NIK:           nik,
		NamaLengkap:   randomName(),
		TempatLahir:   "Morowali",
		TanggalLahir:  time.Now().AddDate(-20-rand.Intn(30), 0, 0),
		JenisKelamin:  []string{"Laki-laki", "Perempuan"}[rand.Intn(2)],
		Agama:         agamaList[rand.Intn(len(agamaList))],
		Pekerjaan:     pekerjaans[rand.Intn(len(pekerjaans))],
		Alamat:        fmt.Sprintf("Desa Bahodopi Dusun %d", rand.Intn(5)+1),
	}
}

func generateLostItem(docID uint, owner string) *models.LostItem {
	tmpl := itemTypes[rand.Intn(len(itemTypes))]
	var desc string
	
	switch tmpl.Name {
	case "KTP":
		nik := fmt.Sprintf("72%014d", rand.Intn(99999999999999))
		desc = fmt.Sprintf(tmpl.DescTemplate, nik, owner)
	case "SIM":
		simNum := fmt.Sprintf("%012d", rand.Intn(999999999999))
		gol := simGolongan[rand.Intn(len(simGolongan))]
		desc = fmt.Sprintf(tmpl.DescTemplate, simNum, gol, owner)
	case "STNK":
		motor := motorBrands[rand.Intn(len(motorBrands))]
		nopol := fmt.Sprintf("DN %d %s", 1000+rand.Intn(9000), string(rune(65+rand.Intn(26)))+string(rune(65+rand.Intn(26))))
		desc = fmt.Sprintf(tmpl.DescTemplate, motor, nopol, owner)
	case "ATM":
		bank := bankNames[rand.Intn(len(bankNames))]
		norek := fmt.Sprintf("%010d", rand.Intn(9999999999))
		desc = fmt.Sprintf(tmpl.DescTemplate, bank, norek, owner)
	case "BPKB":
		mobil := mobilBrands[rand.Intn(len(mobilBrands))]
		noBpkb := fmt.Sprintf("M-%06d", 100000+rand.Intn(900000))
		desc = fmt.Sprintf(tmpl.DescTemplate, mobil, noBpkb, owner)
	case "IJAZAH":
		level := ijazahLevels[rand.Intn(len(ijazahLevels))]
		noIjazah := fmt.Sprintf("IJ/%05d", 10000+rand.Intn(90000))
		desc = fmt.Sprintf(tmpl.DescTemplate, level, noIjazah, owner)
	}
	
	return &models.LostItem{
		LostDocumentID: docID,
		NamaBarang:     tmpl.Name,
		Deskripsi:      desc,
	}
}

func intToRoman(num int) string {
	roman := map[int]string{
		1: "I", 2: "II", 3: "III", 4: "IV", 5: "V", 6: "VI",
		7: "VII", 8: "VIII", 9: "IX", 10: "X", 11: "XI", 12: "XII",
	}
	return roman[num]
}