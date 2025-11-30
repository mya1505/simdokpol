package services

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"simdokpol/internal/config"
	"simdokpol/internal/models"
	"simdokpol/internal/utils"
	"strings"
	"time"

	"gorm.io/gorm"
)

type BackupService interface {
	CreateBackup(actorID uint) (backupPath string, err error)
	RestoreBackup(uploadedFile io.Reader, actorID uint) error
}

type backupService struct {
	db            *gorm.DB
	cfg           *config.Config
	configService ConfigService
	auditService  AuditLogService
}

func NewBackupService(db *gorm.DB, cfg *config.Config, configService ConfigService, auditService AuditLogService) BackupService {
	return &backupService{
		db:            db,
		cfg:           cfg,
		configService: configService,
		auditService:  auditService,
	}
}

func (s *backupService) getCleanDBPath() string {
	dsnParts := strings.Split(s.cfg.DBDSN, "?")
	return dsnParts[0]
}

func (s *backupService) CreateBackup(actorID uint) (string, error) {
	appConfig, err := s.configService.GetConfig()
	if err != nil {
		return "", fmt.Errorf("gagal konfigurasi: %w", err)
	}

	backupDir := appConfig.BackupPath
	if backupDir == "" {
		backupDir = filepath.Join(utils.GetAppDataDir(), "backups")
	}
	os.MkdirAll(backupDir, 0755)

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	destinationPath := filepath.Join(backupDir, fmt.Sprintf("backup-%s.db", timestamp))

	if s.cfg.DBDialect == "sqlite" {
		os.Remove(destinationPath)
		if err := s.db.Exec("VACUUM INTO ?", destinationPath).Error; err != nil {
			return "", fmt.Errorf("gagal SQLite Hot Backup: %w", err)
		}
	} else {
		// Fallback copy manual (hanya jika connection closed, risky)
		return "", fmt.Errorf("backup hanya support SQLite")
	}

	s.auditService.LogActivity(actorID, models.AuditBackupCreated, "Backup: "+filepath.Base(destinationPath))
	return destinationPath, nil
}

func (s *backupService) RestoreBackup(uploadedFile io.Reader, actorID uint) error {
	if s.cfg.DBDialect != "sqlite" {
		return fmt.Errorf("restore hanya support SQLite")
	}

	targetPath := s.getCleanDBPath()
	tempNewPath := targetPath + ".new"
	
	newFile, err := os.Create(tempNewPath)
	if err != nil {
		return fmt.Errorf("gagal buat temp file: %w", err)
	}
	
	_, copyErr := io.Copy(newFile, uploadedFile)
	newFile.Close()
	
	if copyErr != nil {
		os.Remove(tempNewPath)
		return fmt.Errorf("gagal copy data: %w", copyErr)
	}

	backupPath := targetPath + ".bak." + time.Now().Format("20060102150405")
	
	// Rename file asli ke backup (Windows mungkin fail jika terkunci)
	if err := os.Rename(targetPath, backupPath); err != nil {
		os.Remove(tempNewPath)
		if strings.Contains(err.Error(), "process cannot access") {
			return fmt.Errorf("DB TERKUNCI: Tutup aplikasi dan rename file .db secara manual.")
		}
		return fmt.Errorf("gagal backup file lama: %w", err)
	}

	if err := os.Rename(tempNewPath, targetPath); err != nil {
		os.Rename(backupPath, targetPath) // Rollback
		return fmt.Errorf("gagal aktifkan DB baru: %w", err)
	}

	s.auditService.LogActivity(actorID, models.AuditRestoreFromFile, "Restore DB Sukses")
	return nil
}