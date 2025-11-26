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

// ... (Fungsi CreateBackup TIDAK BERUBAH, gunakan yang lama) ...
func (s *backupService) CreateBackup(actorID uint) (string, error) {
	appConfig, err := s.configService.GetConfig()
	if err != nil {
		return "", fmt.Errorf("gagal mendapatkan konfigurasi aplikasi: %w", err)
	}

	backupDir := appConfig.BackupPath
	if backupDir == "" {
		backupDir = filepath.Join(utils.GetAppDataDir(), "backups")
	}

	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("gagal membuat direktori backup di '%s': %w", backupDir, err)
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	destinationPath := filepath.Join(backupDir, fmt.Sprintf("backup-simdokpol-%s.db", timestamp))

	if s.cfg.DBDialect == "sqlite" {
		os.Remove(destinationPath)
		if err := s.db.Exec("VACUUM INTO ?", destinationPath).Error; err != nil {
			return "", fmt.Errorf("gagal melakukan SQLite Hot Backup: %w", err)
		}
	} else {
		sourcePath := s.getCleanDBPath()
		if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
			return "", fmt.Errorf("backup file hanya support SQLite. Gunakan tools DB untuk %s", s.cfg.DBDialect)
		}

		sourceFile, err := os.Open(sourcePath)
		if err != nil {
			return "", fmt.Errorf("gagal membuka file database sumber: %w", err)
		}
		defer sourceFile.Close()

		destinationFile, err := os.Create(destinationPath)
		if err != nil {
			return "", fmt.Errorf("gagal membuat file database backup: %w", err)
		}
		defer destinationFile.Close()

		if _, err := io.Copy(destinationFile, sourceFile); err != nil {
			return "", fmt.Errorf("gagal menyalin data ke file backup: %w", err)
		}
	}

	s.auditService.LogActivity(actorID, models.AuditBackupCreated, fmt.Sprintf("Membuat file backup baru: %s", filepath.Base(destinationPath)))

	return destinationPath, nil
}

// --- FIX RESTORE WINDOWS LOCK ---
func (s *backupService) RestoreBackup(uploadedFile io.Reader, actorID uint) error {
	if s.cfg.DBDialect != "sqlite" {
		return fmt.Errorf("fitur restore hanya tersedia untuk database SQLite")
	}

	targetPath := s.getCleanDBPath()
	
	// 1. Tulis uploaded file ke file temporary dulu (.db.new)
	// Ini menghindari lock karena kita menulis ke file baru, bukan file yang sedang dibuka GORM
	tempNewPath := targetPath + ".new"
	
	newFile, err := os.Create(tempNewPath)
	if err != nil {
		return fmt.Errorf("gagal membuat file temporary: %w", err)
	}
	
	// Pastikan file temp tertutup walau ada error copy
	_, copyErr := io.Copy(newFile, uploadedFile)
	newFile.Close() 
	
	if copyErr != nil {
		os.Remove(tempNewPath) // Bersihkan sampah
		return fmt.Errorf("gagal menyalin data upload: %w", copyErr)
	}

	// 2. Lakukan Swap File
	// Strategy: Rename DB Aktif -> .bak, Rename .new -> DB Aktif
	// Windows mengizinkan rename file yang sedang terbuka (biasanya), tapi tidak delete.
	
	backupPath := targetPath + ".bak." + time.Now().Format("20060102150405")
	
	// Rename file asli ke backup
	if err := os.Rename(targetPath, backupPath); err != nil {
		os.Remove(tempNewPath)
		// Pesan error yang membantu user Windows
		return fmt.Errorf("FILE TERKUNCI (Windows Lock). Tutup aplikasi lalu rename manual file '%s' menjadi '%s', dan '%s' menjadi '%s'", 
			filepath.Base(targetPath), filepath.Base(backupPath), filepath.Base(tempNewPath), filepath.Base(targetPath))
	}

	// Rename file baru ke nama asli
	if err := os.Rename(tempNewPath, targetPath); err != nil {
		// Critical: Coba kembalikan file lama jika gagal
		os.Rename(backupPath, targetPath)
		return fmt.Errorf("gagal mengaktifkan database baru: %w", err)
	}

	s.auditService.LogActivity(actorID, models.AuditRestoreFromFile, "Database dipulihkan (Swap Method). Restart aplikasi diperlukan.")

	return nil
}
