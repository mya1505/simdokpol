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
	MigrateDataTo(targetConfig dto.DBTestRequest, actorID uint) error
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

func (s *dataMigrationService) MigrateDataTo(target dto.DBTestRequest, actorID uint) error {
	// 1. Buka Koneksi ke Target Database
	targetDB, err := s.openTargetConnection(target)
	if err != nil {
		return fmt.Errorf("gagal koneksi ke target database: %w", err)
	}
	
	// 2. Auto Migrate Target (Buat Tabel)
	log.Println("MIGRASI: Membuat skema tabel di target...")
	err = targetDB.AutoMigrate(
		&models.Configuration{}, &models.User{}, &models.Resident{},
		&models.LostDocument{}, &models.LostItem{}, &models.AuditLog{},
		&models.ItemTemplate{}, &models.License{},
	)
	if err != nil {
		return fmt.Errorf("gagal membuat tabel di target: %w", err)
	}

	// 3. Mulai Transaksi Migrasi Data
	// FIX QF1003: Gunakan switch case
	switch target.DBDialect {
	case "mysql":
		targetDB.Exec("SET FOREIGN_KEY_CHECKS = 0")
		defer targetDB.Exec("SET FOREIGN_KEY_CHECKS = 1")
	case "sqlite":
		targetDB.Exec("PRAGMA foreign_keys = OFF")
		defer targetDB.Exec("PRAGMA foreign_keys = ON")
	}
	// Postgres tidak perlu disable global FK untuk urutan insert ini

	log.Println("MIGRASI: Memindahkan data Configuration...")
	if err := s.copyTable(s.currentDB, targetDB, &models.Configuration{}); err != nil { return err }

	log.Println("MIGRASI: Memindahkan data License...")
	if err := s.copyTable(s.currentDB, targetDB, &models.License{}); err != nil { return err }
	
	log.Println("MIGRASI: Memindahkan data ItemTemplate...")
	if err := s.copyTable(s.currentDB, targetDB, &models.ItemTemplate{}); err != nil { return err }

	log.Println("MIGRASI: Memindahkan data User...")
	if err := s.copyTable(s.currentDB, targetDB, &models.User{}); err != nil { return err }

	log.Println("MIGRASI: Memindahkan data Resident...")
	if err := s.copyTable(s.currentDB, targetDB, &models.Resident{}); err != nil { return err }

	log.Println("MIGRASI: Memindahkan data LostDocument...")
	if err := s.copyTable(s.currentDB, targetDB, &models.LostDocument{}); err != nil { return err }

	log.Println("MIGRASI: Memindahkan data LostItem...")
	if err := s.copyTable(s.currentDB, targetDB, &models.LostItem{}); err != nil { return err }

	log.Println("MIGRASI: Memindahkan data AuditLog...")
	if err := s.copyTable(s.currentDB, targetDB, &models.AuditLog{}); err != nil { return err }

	s.auditService.LogActivity(actorID, "MIGRASI DATABASE", fmt.Sprintf("Data berhasil disalin ke database baru (%s)", target.DBDialect))

	return nil
}

func (s *dataMigrationService) copyTable(src, dest *gorm.DB, model interface{}) error {
	return src.Model(model).FindInBatches(model, 100, func(tx *gorm.DB, batch int) error {
		return dest.Clauses(clause.OnConflict{DoNothing: true}).Create(tx.Statement.Dest).Error
	}).Error
}

func (s *dataMigrationService) openTargetConnection(req dto.DBTestRequest) (*gorm.DB, error) {
	var dsn string
	var dialector gorm.Dialector

	switch req.DBDialect {
	case "mysql":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			req.DBUser, req.DBPass, req.DBHost, req.DBPort, req.DBName)
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