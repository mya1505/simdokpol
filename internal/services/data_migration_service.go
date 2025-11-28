package services

import (
	"fmt"
	"log"
	"simdokpol/internal/dto"
	"simdokpol/internal/models"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

type DataMigrationService interface {
	// Update signature: Tambah channel untuk progress report
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
	// Helper kirim progress
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
	
	report("schema", 10, "Membuat struktur tabel...")
	err = targetDB.AutoMigrate(
		&models.Configuration{}, &models.User{}, &models.Resident{},
		&models.LostDocument{}, &models.LostItem{}, &models.AuditLog{},
		&models.ItemTemplate{}, &models.License{},
	)
	if err != nil {
		return fmt.Errorf("gagal membuat tabel di target: %w", err)
	}

	// --- MATIKAN FOREIGN KEY CHECKS (CRITICAL FIX) ---
	// Ini mencegah error FK saat insert data yang urutannya mungkin acak atau circular
	switch target.DBDialect {
	case "mysql":
		targetDB.Exec("SET FOREIGN_KEY_CHECKS = 0")
		defer targetDB.Exec("SET FOREIGN_KEY_CHECKS = 1")
	case "sqlite":
		targetDB.Exec("PRAGMA foreign_keys = OFF")
		defer targetDB.Exec("PRAGMA foreign_keys = ON")
	case "postgres":
		// Superuser required. Jika gagal, kita andalkan urutan insert yang benar.
		// session_replication_role = replica menonaktifkan trigger dan FK check
		targetDB.Exec("SET session_replication_role = 'replica'")
		defer targetDB.Exec("SET session_replication_role = 'origin'")
	}

	// --- PROSES SALIN DATA (FIX: UNSCOPED) ---
	// Kita gunakan Unscoped() agar data soft-deleted (DeletedAt != NULL) tetap tersalin.
	// Ini mencegah error FK jika dokumen mereferensi user yang sudah dihapus.

	tables := []struct {
		Name  string
		Model interface{}
		Pct   int
	}{
		{"Konfigurasi", &models.Configuration{}, 15},
		{"Lisensi", &models.License{}, 20},
		{"Template Barang", &models.ItemTemplate{}, 25},
		{"Pengguna", &models.User{}, 35}, // User dulu
		{"Penduduk", &models.Resident{}, 45}, // Lalu Resident
		{"Dokumen", &models.LostDocument{}, 60}, // Baru Dokumen (refer ke User & Resident)
		{"Barang Hilang", &models.LostItem{}, 80},
		{"Log Audit", &models.AuditLog{}, 90},
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

func (s *dataMigrationService) copyTable(src, dest *gorm.DB, model interface{}) error {
	// FIX: Tambahkan Unscoped() agar data terhapus tetap tersalin
	return src.Unscoped().Model(model).FindInBatches(model, 100, func(tx *gorm.DB, batch int) error {
		// Gunakan Clauses OnConflict DoNothing agar idempotency terjaga
		return dest.Clauses(clause.OnConflict{DoNothing: true}).Create(tx.Statement.Dest).Error
	}).Error
}

func (s *dataMigrationService) openTargetConnection(req dto.DBTestRequest) (*gorm.DB, error) {
	var dsn string
	var dialector gorm.Dialector

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