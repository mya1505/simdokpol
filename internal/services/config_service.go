package services

import (
	"log"
	"os"
	"path/filepath"
	"simdokpol/internal/dto"
	"simdokpol/internal/repositories"
	"simdokpol/internal/utils"
	"strconv"
	"strings"
	"sync" // <-- FIX B-03: Import sync
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
	cachedLocation *time.Location
	cachedConfig   *dto.AppConfig
	mu             sync.RWMutex // <-- FIX B-03: Mutex untuk thread-safety
}

func NewConfigService(configRepo repositories.ConfigRepository) ConfigService {
	return &configService{configRepo: configRepo}
}

func (s *configService) SaveConfig(configData map[string]string) error {
	// FIX B-03: Lock mutex saat menulis/menghapus cache
	s.mu.Lock()
	s.cachedLocation = nil
	s.cachedConfig = nil
	s.mu.Unlock()

	if err := s.configRepo.SetMultiple(nil, configData); err != nil {
		return err
	}

	envUpdates := make(map[string]string)
	dbKeys := map[string]string{
		"db_dialect": "DB_DIALECT",
		"db_host":    "DB_HOST",
		"db_port":    "DB_PORT",
		"db_name":    "DB_NAME",
		"db_user":    "DB_USER",
		"db_pass":    "DB_PASS",
		"db_dsn":     "DB_DSN",
	}

	hasEnvChanges := false
	for jsonKey, envKey := range dbKeys {
		if val, ok := configData[jsonKey]; ok {
			envUpdates[envKey] = val
			hasEnvChanges = true
		}
	}

	if hasEnvChanges {
		log.Println("INFO: Mendeteksi perubahan konfigurasi database, memperbarui file .env...")
		if err := utils.UpdateEnvFile(envUpdates); err != nil {
			log.Printf("ERROR: Gagal memperbarui file .env: %v", err)
		}
	}

	return nil
}

func (s *configService) GetLocation() (*time.Location, error) {
	// FIX B-03: Read Lock untuk cek cache
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

	// FIX B-03: Write Lock untuk update cache
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

	archiveDays, _ := strconv.Atoi(allConfigs["archive_duration_days"])
	backupPath := allConfigs["backup_path"]
	if backupPath == "" {
		backupPath = filepath.Join(utils.GetAppDataDir(), "backups")
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

		DBDialect:     allConfigs["db_dialect"],
		DBHost:        allConfigs["db_host"],
		DBPort:        allConfigs["db_port"],
		DBUser:        allConfigs["db_user"],
		DBName:        allConfigs["db_name"],
		DBDSN:         allConfigs["db_dsn"],
		DBSSLMode:     allConfigs["db_sslmode"], // <-- BACA DARI DB
		LicenseStatus: allConfigs["license_status"],
	}

	// Fallback ke Environment Variable jika di DB kosong
	if appConfig.DBDialect == "" {
		appConfig.DBDialect = strings.ToLower(os.Getenv("DB_DIALECT"))
		if appConfig.DBDialect == "" { appConfig.DBDialect = "sqlite" }
	}
	if appConfig.DBHost == "" { appConfig.DBHost = os.Getenv("DB_HOST") }
	if appConfig.DBPort == "" { appConfig.DBPort = os.Getenv("DB_PORT") }
	if appConfig.DBUser == "" { appConfig.DBUser = os.Getenv("DB_USER") }
	if appConfig.DBName == "" { appConfig.DBName = os.Getenv("DB_NAME") }
	
	// Logic SSL Mode
	if appConfig.DBSSLMode == "" {
		appConfig.DBSSLMode = os.Getenv("DB_SSLMODE")
		if appConfig.DBSSLMode == "" { appConfig.DBSSLMode = "disable" }
	}

	if appConfig.DBDSN == "" {
		appConfig.DBDSN = os.Getenv("DB_DSN")
		if appConfig.DBDialect == "sqlite" && appConfig.DBDSN == "" {
			appConfig.DBDSN = "simdokpol.db?_foreign_keys=on"
		}
	}

	s.mu.Lock()
	s.cachedConfig = appConfig
	s.mu.Unlock()
	
	return appConfig, nil
}
