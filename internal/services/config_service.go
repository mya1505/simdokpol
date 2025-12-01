package services

import (
	"log"
	"os"
	"path/filepath"
	"simdokpol/internal/dto"
	"simdokpol/internal/models"
	"simdokpol/internal/repositories"
	"simdokpol/internal/utils"
	"strconv"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
)

const IsSetupCompleteKey = "is_setup_complete"

type ConfigService interface {
	IsSetupComplete() (bool, error)
	GetConfig() (*dto.AppConfig, error)
	SaveConfig(configData map[string]string) error
	GetLocation() (*time.Location, error)
}

type configService struct {
	configRepo     repositories.ConfigRepository
	db             *gorm.DB
	cachedLocation *time.Location
	cachedConfig   *dto.AppConfig
	mu             sync.RWMutex
}

func NewConfigService(configRepo repositories.ConfigRepository, db *gorm.DB) ConfigService {
	return &configService{
		configRepo: configRepo,
		db:         db,
	}
}

func (s *configService) SaveConfig(configData map[string]string) error {
	s.mu.Lock()
	s.cachedLocation = nil
	s.cachedConfig = nil
	s.mu.Unlock()

	// 1. Simpan ke Database
	if err := s.configRepo.SetMultiple(nil, configData); err != nil {
		return err
	}

	// 2. Simpan ke .env (Mapping agar sesuai nama environment variable)
	envUpdates := make(map[string]string)
	envKeyMapping := map[string]string{
		"DB_DIALECT": "DB_DIALECT",
		"DB_DSN":     "DB_DSN",
		"DB_HOST":    "DB_HOST",
		"DB_PORT":    "DB_PORT",
		"DB_USER":    "DB_USER",
		"DB_PASS":    "DB_PASS",
		"DB_NAME":    "DB_NAME",
		"DB_SSLMODE": "DB_SSLMODE",
		
		"enable_https":    "ENABLE_HTTPS",
		"session_timeout": "SESSION_TIMEOUT",
		"idle_timeout":    "IDLE_TIMEOUT",
		
		// PENTING: Flag setup juga harus masuk ke .env
		IsSetupCompleteKey: "IS_SETUP_COMPLETE",
	}

	hasEnvChanges := false
	for configKey, envKey := range envKeyMapping {
		// Coba cari exact match atau lowercase match
		val, ok := configData[configKey]
		if !ok {
			val, ok = configData[strings.ToLower(configKey)]
		}

		if ok {
			envUpdates[envKey] = val
			hasEnvChanges = true
		}
	}

	if hasEnvChanges {
		log.Println("INFO: Memperbarui file .env dengan konfigurasi baru...")
		if err := utils.UpdateEnvFile(envUpdates); err != nil {
			log.Printf("ERROR: Gagal memperbarui file .env: %v", err)
			// Tidak return error karena DB sudah tersimpan
		} else {
			log.Println("SUCCESS: File .env berhasil diperbarui")
		}
	}

	return nil
}

func (s *configService) GetLocation() (*time.Location, error) {
	s.mu.RLock()
	if s.cachedLocation != nil {
		defer s.mu.RUnlock()
		return s.cachedLocation, nil
	}
	s.mu.RUnlock()

	config, err := s.GetConfig()
	if err != nil {
		return time.UTC, err
	}

	var loc *time.Location
	if config.ZonaWaktu == "" {
		loc = time.UTC
	} else {
		loc, err = time.LoadLocation(config.ZonaWaktu)
		if err != nil {
			return time.UTC, err
		}
	}

	s.mu.Lock()
	s.cachedLocation = loc
	s.mu.Unlock()

	return loc, nil
}

func (s *configService) IsSetupComplete() (bool, error) {
	config, err := s.configRepo.Get(IsSetupCompleteKey)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	return config.Value == "true", nil
}

func (s *configService) GetConfig() (*dto.AppConfig, error) {
	s.mu.RLock()
	if s.cachedConfig != nil {
		defer s.mu.RUnlock()
		return s.cachedConfig, nil
	}
	s.mu.RUnlock()

	allConfigs, err := s.configRepo.GetAll()
	if err != nil {
		return nil, err
	}

	// Sync Nomor Surat (opsional)
	if s.db != nil {
		var lastDoc models.LostDocument
		if err := s.db.Order("id desc").First(&lastDoc).Error; err == nil {
			parts := strings.Split(lastDoc.NomorSurat, "/")
			if len(parts) > 1 {
				allConfigs["nomor_surat_terakhir"] = parts[1]
			}
		}
	}

	archiveDays, _ := strconv.Atoi(allConfigs["archive_duration_days"])
	backupPath := allConfigs["backup_path"]
	if backupPath == "" {
		backupPath = filepath.Join(utils.GetAppDataDir(), "backups")
	}

	enableHttpsDB := allConfigs["enable_https"]
	enableHttpsEnv := os.Getenv("ENABLE_HTTPS")
	isHttps := enableHttpsDB == "true" || (enableHttpsDB == "" && enableHttpsEnv == "true")

	sessionTimeout, _ := strconv.Atoi(allConfigs["session_timeout"])
	if sessionTimeout == 0 {
		sessionTimeout, _ = strconv.Atoi(os.Getenv("SESSION_TIMEOUT"))
		if sessionTimeout == 0 {
			sessionTimeout = 480
		}
	}

	idleTimeout, _ := strconv.Atoi(allConfigs["idle_timeout"])
	if idleTimeout == 0 {
		idleTimeout, _ = strconv.Atoi(os.Getenv("IDLE_TIMEOUT"))
		if idleTimeout == 0 {
			idleTimeout = 15
		}
	}

	appConfig := &dto.AppConfig{
		IsSetupComplete:     allConfigs[IsSetupCompleteKey] == "true",
		KopBaris1:           allConfigs["kop_baris_1"],
		KopBaris2:           allConfigs["kop_baris_2"],
		KopBaris3:           allConfigs["kop_baris_3"],
		NamaKantor:          allConfigs["nama_kantor"],
		TempatSurat:         allConfigs["tempat_surat"],
		FormatNomorSurat:    allConfigs["format_nomor_surat"],
		NomorSuratTerakhir:  allConfigs["nomor_surat_terakhir"],
		ZonaWaktu:           allConfigs["zona_waktu"],
		BackupPath:          backupPath,
		ArchiveDurationDays: archiveDays,

		SessionTimeout: sessionTimeout,
		IdleTimeout:    idleTimeout,
		EnableHTTPS:    isHttps,

		DBDialect:     allConfigs["db_dialect"],
		DBHost:        allConfigs["db_host"],
		DBPort:        allConfigs["db_port"],
		DBUser:        allConfigs["db_user"],
		DBName:        allConfigs["db_name"],
		DBDSN:         allConfigs["db_dsn"],
		DBSSLMode:     allConfigs["db_sslmode"],
		LicenseStatus: allConfigs["license_status"],
	}

	if appConfig.DBDialect == "" {
		appConfig.DBDialect = strings.ToLower(os.Getenv("DB_DIALECT"))
		if appConfig.DBDialect == "" {
			appConfig.DBDialect = "sqlite"
		}
	}
	if appConfig.DBHost == "" { appConfig.DBHost = os.Getenv("DB_HOST") }
	if appConfig.DBPort == "" { appConfig.DBPort = os.Getenv("DB_PORT") }
	if appConfig.DBUser == "" { appConfig.DBUser = os.Getenv("DB_USER") }
	if appConfig.DBName == "" { appConfig.DBName = os.Getenv("DB_NAME") }
	if appConfig.DBSSLMode == "" { appConfig.DBSSLMode = os.Getenv("DB_SSLMODE") }
	
	// DSN Fallback & Normalization
	if appConfig.DBDSN == "" {
		appConfig.DBDSN = os.Getenv("DB_DSN")
	}
	if appConfig.DBDialect == "sqlite" {
		if appConfig.DBDSN == "" {
			appConfig.DBDSN = "simdokpol.db?_foreign_keys=on"
		}
		// Ensure Absolute Path in Config Object
		cleanPath := strings.TrimPrefix(appConfig.DBDSN, "file:")
		cleanPath = strings.Split(cleanPath, "?")[0]
		if !filepath.IsAbs(cleanPath) {
			appConfig.DBDSN = filepath.Join(utils.GetAppDataDir(), cleanPath) + "?_foreign_keys=on"
		}
	}

	s.mu.Lock()
	s.cachedConfig = appConfig
	s.mu.Unlock()

	return appConfig, nil
}