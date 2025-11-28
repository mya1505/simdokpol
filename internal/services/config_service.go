package services

import (
	"fmt"
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
	db             *gorm.DB // <-- FIELD BARU: Untuk Sync Nomor Surat
	cachedLocation *time.Location
	cachedConfig   *dto.AppConfig
	mu             sync.RWMutex
}

// NewConfigService sekarang menerima *gorm.DB
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

	if err := s.configRepo.SetMultiple(nil, configData); err != nil {
		return err
	}

	envUpdates := make(map[string]string)
	dbKeys := map[string]string{
		"db_dialect":      "DB_DIALECT",
		"db_host":         "DB_HOST",
		"db_port":         "DB_PORT",
		"db_name":         "DB_NAME",
		"db_user":         "DB_USER",
		"db_pass":         "DB_PASS",
		"db_dsn":          "DB_DSN",
		"db_sslmode":      "DB_SSLMODE",
		"enable_https":    "ENABLE_HTTPS",
		"session_timeout": "SESSION_TIMEOUT",
		"idle_timeout":    "IDLE_TIMEOUT",
	}

	hasEnvChanges := false
	for jsonKey, envKey := range dbKeys {
		if val, ok := configData[jsonKey]; ok {
			envUpdates[envKey] = val
			hasEnvChanges = true
		}
	}

	if hasEnvChanges {
		log.Println("INFO: Memperbarui file .env dengan konfigurasi baru...")
		if err := utils.UpdateEnvFile(envUpdates); err != nil {
			log.Printf("ERROR: Gagal memperbarui file .env: %v", err)
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

	// --- FIX BUG 3: SINKRONISASI NOMOR SURAT TERAKHIR ---
	// Ambil nomor urut dari dokumen terakhir di DB
	// Logic ini memastikan angka di Settings selalu sesuai realita
	if s.db != nil {
		var lastDoc models.LostDocument
		// Ambil dokumen terakhir berdasarkan ID
		if err := s.db.Order("id desc").First(&lastDoc).Error; err == nil {
			// Parsing nomor surat (Format: SKH/001/...)
			parts := strings.Split(lastDoc.NomorSurat, "/")
			if len(parts) > 1 {
				// Jika berhasil diparse, update nilai di map config sementara (untuk ditampilkan)
				// Kita tidak simpan ke DB config dulu biar gak spam write, cukup di view saja
				allConfigs["nomor_surat_terakhir"] = parts[1] 
			}
		}
	}
	// ----------------------------------------------------

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
		if sessionTimeout == 0 { sessionTimeout = 480 }
	}

	idleTimeout, _ := strconv.Atoi(allConfigs["idle_timeout"])
	if idleTimeout == 0 {
		idleTimeout, _ = strconv.Atoi(os.Getenv("IDLE_TIMEOUT"))
		if idleTimeout == 0 { idleTimeout = 15 }
	}

	appConfig := &dto.AppConfig{
		IsSetupComplete:     allConfigs[IsSetupCompleteKey] == "true",
		KopBaris1:           allConfigs["kop_baris_1"],
		KopBaris2:           allConfigs["kop_baris_2"],
		KopBaris3:           allConfigs["kop_baris_3"],
		NamaKantor:          allConfigs["nama_kantor"],
		TempatSurat:         allConfigs["tempat_surat"],
		FormatNomorSurat:    allConfigs["format_nomor_surat"],
		NomorSuratTerakhir:  allConfigs["nomor_surat_terakhir"], // Sudah disync di atas
		ZonaWaktu:           allConfigs["zona_waktu"],
		BackupPath:          backupPath,
		ArchiveDurationDays: archiveDays,
		
		SessionTimeout:      sessionTimeout,
		IdleTimeout:         idleTimeout,
		EnableHTTPS:         isHttps,

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
		if appConfig.DBDialect == "" { appConfig.DBDialect = "sqlite" }
	}
	if appConfig.DBHost == "" { appConfig.DBHost = os.Getenv("DB_HOST") }
	if appConfig.DBPort == "" { appConfig.DBPort = os.Getenv("DB_PORT") }
	if appConfig.DBUser == "" { appConfig.DBUser = os.Getenv("DB_USER") }
	if appConfig.DBName == "" { appConfig.DBName = os.Getenv("DB_NAME") }
	if appConfig.DBSSLMode == "" { 
		appConfig.DBSSLMode = os.Getenv("DB_SSLMODE") 
		if appConfig.DBSSLMode == "" { appConfig.DBSSLMode = "disable" }
	}

	// --- FIX BUG 5: Path DSN Kosong/Relative ---
	if appConfig.DBDSN == "" {
		appConfig.DBDSN = os.Getenv("DB_DSN")
	}
	// Pastikan path SQLite selalu absolut ke AppData agar tidak "hilang"
	if appConfig.DBDialect == "sqlite" {
		if appConfig.DBDSN == "" || !filepath.IsAbs(strings.Split(appConfig.DBDSN, "?")[0]) {
			fname := "simdokpol.db"
			if appConfig.DBDSN != "" {
				// Ambil nama file dari DSN lama jika ada
				cleanDSN := strings.TrimPrefix(appConfig.DBDSN, "file:")
				parts := strings.Split(cleanDSN, "?")
				if parts[0] != "" { fname = filepath.Base(parts[0]) }
			}
			// Rebuild absolute path
			appConfig.DBDSN = filepath.Join(utils.GetAppDataDir(), fmt.Sprintf("%s?_foreign_keys=on", fname))
		}
	}
	// -------------------------------------------

	s.mu.Lock()
	s.cachedConfig = appConfig
	s.mu.Unlock()
	
	return appConfig, nil
}