package services

import (
	"fmt"
	"simdokpol/internal/dto"
	"simdokpol/internal/models"
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

type DataMigrationService interface {
	MigrateDataTo(targetConfig dto.DBTestRequest, actorID uint, progressChan chan<- dto.MigrationProgress) error
}

type dataMigrationService struct {
	currentDB     *gorm.DB
	auditService  AuditLogService
	configService ConfigService
}

func NewDataMigrationService(currentDB *gorm.DB, auditService AuditLogService, configService ConfigService) DataMigrationService {
	return &dataMigrationService{
		currentDB:     currentDB,
		auditService:  auditService,
		configService: configService,
	}
}

func (s *dataMigrationService) MigrateDataTo(target dto.DBTestRequest, actorID uint, progressChan chan<- dto.MigrationProgress) error {
	report := func(step string, pct int, msg string) {
		if progressChan != nil {
			progressChan <- dto.MigrationProgress{Step: step, Percent: pct, Message: msg}
		}
	}

	report("connect", 5, "Menghubungkan ke database target...")
	targetDB, err := s.openTargetConnection(target)
	if err != nil {
		return fmt.Errorf("gagal koneksi ke target database: %w", err)
	}
	
	sqlDB, _ := targetDB.DB()
	defer sqlDB.Close()

	report("schema", 10, "Membuat struktur tabel...")
	// AutoMigrate akan membuat tabel di target
	err = targetDB.AutoMigrate(
		&models.Configuration{}, &models.User{}, &models.Resident{},
		&models.LostDocument{}, &models.LostItem{}, &models.AuditLog{},
		&models.ItemTemplate{}, &models.License{},
	)
	if err != nil {
		return fmt.Errorf("gagal membuat tabel di target: %w", err)
	}

	// Coba Matikan FK Checks (Best Effort - Tergantung Permission User DB)
	switch target.DBDialect {
	case "mysql":
		targetDB.Exec("SET FOREIGN_KEY_CHECKS = 0")
		defer targetDB.Exec("SET FOREIGN_KEY_CHECKS = 1")
	case "sqlite":
		targetDB.Exec("PRAGMA foreign_keys = OFF")
		defer targetDB.Exec("PRAGMA foreign_keys = ON")
	case "postgres":
		// Ini butuh superuser. Jika gagal, kita andalkan urutan insert yang benar.
		targetDB.Exec("SET session_replication_role = 'replica'")
		defer targetDB.Exec("SET session_replication_role = 'origin'")
	}

	// --- FIX UTAMA: GUNAKAN SLICE, BUKAN STRUCT TUNGGAL ---
	// GORM butuh slice pointer (&[]Model{}) agar FindInBatches bekerja benar.
	tables := []struct {
		Name  string
		Model interface{} // Pointer to Slice
		Pct   int
	}{
		{"Konfigurasi", &[]models.Configuration{}, 15},
		{"Lisensi", &[]models.License{}, 20},
		{"Template Barang", &[]models.ItemTemplate{}, 25},
		
		// Urutan SANGAT PENTING agar FK valid:
		{"Pengguna", &[]models.User{}, 35},        // User dulu (referensi oleh Doc)
		{"Penduduk", &[]models.Resident{}, 45},    // Resident dulu (referensi oleh Doc)
		{"Dokumen", &[]models.LostDocument{}, 60}, // Baru Dokumen
		{"Barang Hilang", &[]models.LostItem{}, 80}, // Item referensi Doc
		{"Log Audit", &[]models.AuditLog{}, 90},
	}

	for _, t := range tables {
		report(t.Name, t.Pct, fmt.Sprintf("Menyalin data %s...", t.Name))
		if err := s.copyTable(s.currentDB, targetDB, t.Model); err != nil {
			return fmt.Errorf("gagal menyalin %s: %w", t.Name, err)
		}
	}

	report("finish", 100, "Finalisasi migrasi...")
	s.auditService.LogActivity(actorID, models.AuditSettingsUpdated, fmt.Sprintf("Data berhasil disalin ke database baru (%s)", target.DBDialect))

	return nil
}

func (s *dataMigrationService) copyTable(src, dest *gorm.DB, modelSlice interface{}) error {
	// Gunakan Unscoped() agar data soft-deleted tetap tersalin (mencegah FK error jika referensi ke user yg dihapus)
	// Model(modelSlice) memberitahu GORM tabel mana yang dipakai.
	// FindInBatches(modelSlice) mengisi slice tersebut.
	
	return src.Unscoped().Model(modelSlice).FindInBatches(modelSlice, 100, func(tx *gorm.DB, batch int) error {
		// Insert Batch ke Target
		// Clause OnConflict DoNothing untuk idempotency (kalau diulang tidak double)
		return dest.Clauses(clause.OnConflict{DoNothing: true}).Create(tx.Statement.Dest).Error
	}).Error
}

func (s *dataMigrationService) openTargetConnection(req dto.DBTestRequest) (*gorm.DB, error) {
	var dsn string
	var dialector gorm.Dialector

	// Auto-Clean Host:Port input
	if strings.Contains(req.DBHost, ":") {
		parts := strings.Split(req.DBHost, ":")
		if len(parts) == 2 {
			req.DBHost = parts[0]
			req.DBPort = parts[1]
		}
	}

	switch req.DBDialect {
	case "mysql":
		tlsOption := "false"
		if req.DBSSLMode == "require" { tlsOption = "skip-verify" } else if req.DBSSLMode == "verify-full" { tlsOption = "true" }
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&tls=%s",
			req.DBUser, req.DBPass, req.DBHost, req.DBPort, req.DBName, tlsOption)
		dialector = mysql.Open(dsn)
	case "postgres":
		sslMode := req.DBSSLMode
		if sslMode == "" { sslMode = "disable" }
		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Jakarta",
			req.DBHost, req.DBUser, req.DBPass, req.DBName, req.DBPort, sslMode)
		dialector = postgres.Open(dsn)
	case "sqlite":
		dsn = req.DBName 
		if dsn == "" { dsn = "migrated_simdokpol.db?_foreign_keys=on" }
		dialector = sqlite.Open(dsn)
	default:
		return nil, fmt.Errorf("dialek tidak didukung")
	}

	return gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
}