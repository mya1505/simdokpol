package controllers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"simdokpol/internal/dto"
	"simdokpol/internal/models"
	"simdokpol/internal/services"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type ConfigController struct {
	configService    services.ConfigService
	userService      services.UserService
	backupService    services.BackupService
	migrationService services.DataMigrationService
}

func NewConfigController(
	configService services.ConfigService,
	userService services.UserService,
	backupService services.BackupService,
	migrationService services.DataMigrationService,
) *ConfigController {
	return &ConfigController{
		configService:    configService,
		userService:      userService,
		backupService:    backupService,
		migrationService: migrationService,
	}
}

// --- METHOD BARU ---
// @Summary Ambil Batasan Konfigurasi (Publik)
// @Router /api/config/limits [get]
func (c *ConfigController) GetLimits(ctx *gin.Context) {
	cfg, err := c.configService.GetConfig()
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"session_timeout": 480, "idle_timeout": 15})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"session_timeout": cfg.SessionTimeout,
		"idle_timeout":    cfg.IdleTimeout,
	})
}
// -------------------

type SaveSetupRequest struct {
	DBDialect string `json:"db_dialect" binding:"required"`
	DBDSN     string `json:"db_dsn"`
	DBHost    string `json:"db_host"`
	DBPort    string `json:"db_port"`
	DBUser    string `json:"db_user"`
	DBPass    string `json:"db_pass"`
	DBName    string `json:"db_name"`
	DBSSLMode string `json:"db_sslmode"`
	
	KopBaris1   string `json:"kop_baris_1" binding:"required"`
	KopBaris2   string `json:"kop_baris_2" binding:"required"`
	KopBaris3   string `json:"kop_baris_3" binding:"required"`
	NamaKantor  string `json:"nama_kantor" binding:"required"`
	TempatSurat string `json:"tempat_surat" binding:"required"`

	FormatNomorSurat    string `json:"format_nomor_surat" binding:"required"`
	NomorSuratTerakhir  string `json:"nomor_surat_terakhir" binding:"required"`
	ZonaWaktu           string `json:"zona_waktu" binding:"required"`
	ArchiveDurationDays string `json:"archive_duration_days" binding:"required"`

	AdminNamaLengkap string `json:"admin_nama_lengkap" binding:"required"`
	AdminNRP         string `json:"admin_nrp" binding:"required"`
	AdminPangkat     string `json:"admin_pangkat" binding:"required"`
	AdminPassword    string `json:"admin_password" binding:"required,min=8"`
}

func (c *ConfigController) connectToTargetDB(req SaveSetupRequest) (*gorm.DB, error) {
	var dsn string
	var dialector gorm.Dialector
	switch req.DBDialect {
	case "mysql":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", req.DBUser, req.DBPass, req.DBHost, req.DBPort, req.DBName)
		dialector = mysql.Open(dsn)
	case "postgres":
		sslMode := req.DBSSLMode
		if sslMode == "" { sslMode = "disable" }
		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Jakarta", req.DBHost, req.DBUser, req.DBPass, req.DBName, req.DBPort, sslMode)
		dialector = postgres.Open(dsn)
	default: 
		dsn = req.DBDSN
		if dsn == "" { dsn = "simdokpol.db?_foreign_keys=on" }
		dialector = sqlite.Open(dsn)
	}
	return gorm.Open(dialector, &gorm.Config{Logger: logger.Default.LogMode(logger.Warn)})
}

func (c *ConfigController) SaveSetup(ctx *gin.Context) {
	isSetup, _ := c.configService.IsSetupComplete()
	if isSetup { APIError(ctx, http.StatusForbidden, "Aplikasi sudah dikonfigurasi."); return }
	var req SaveSetupRequest
	if err := ctx.ShouldBindJSON(&req); err != nil { APIError(ctx, http.StatusBadRequest, "Input tidak valid: "+err.Error()); return }
	
	targetDB, err := c.connectToTargetDB(req)
	if err != nil { log.Printf("ERROR: Koneksi target: %v", err); APIError(ctx, http.StatusInternalServerError, "Gagal koneksi DB tujuan."); return }
	if err := targetDB.AutoMigrate(&models.User{}); err != nil { APIError(ctx, http.StatusInternalServerError, "Gagal migrasi tabel user."); return }
	
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.AdminPassword), 10)
	superAdmin := &models.User{NamaLengkap: req.AdminNamaLengkap, NRP: req.AdminNRP, Pangkat: req.AdminPangkat, KataSandi: string(hashedPassword), Peran: models.RoleSuperAdmin, Jabatan: models.RoleSuperAdmin, Regu: "-"}
	if err := targetDB.Where("nrp = ?", superAdmin.NRP).FirstOrCreate(superAdmin).Error; err != nil { APIError(ctx, http.StatusInternalServerError, "Gagal membuat Admin."); return }
	sqlDB, _ := targetDB.DB(); sqlDB.Close()

	configData := map[string]string{
		"DB_DIALECT": req.DBDialect, "DB_DSN": req.DBDSN, "DB_HOST": req.DBHost, "DB_PORT": req.DBPort, "DB_USER": req.DBUser, "DB_PASS": req.DBPass, "DB_NAME": req.DBName, "DB_SSLMODE": req.DBSSLMode,
		"kop_baris_1": req.KopBaris1, "kop_baris_2": req.KopBaris2, "kop_baris_3": req.KopBaris3, "nama_kantor": req.NamaKantor, "tempat_surat": req.TempatSurat,
		"format_nomor_surat": req.FormatNomorSurat, "nomor_surat_terakhir": req.NomorSuratTerakhir, "zona_waktu": req.ZonaWaktu, "archive_duration_days": req.ArchiveDurationDays,
		services.IsSetupCompleteKey: "true",
	}
	if req.DBDialect == "sqlite" && req.DBDSN == "" { configData["DB_DSN"] = "simdokpol.db?_foreign_keys=on" }
	
	if err := c.configService.SaveConfig(configData); err != nil {
		APIError(ctx, http.StatusInternalServerError, "Gagal menyimpan konfigurasi.")
		return
	}
	APIResponse(ctx, http.StatusOK, "Setup berhasil.", nil)
}

// Update fungsi MigrateDatabase
// @Summary Migrasi Data (Stream)
// @Router /api/settings/migrate [post]
func (c *ConfigController) MigrateDatabase(ctx *gin.Context) {
	var req dto.DBTestRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		APIError(ctx, http.StatusBadRequest, "Input konfigurasi tidak valid.")
		return
	}

	// Setup SSE Headers
	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")
	ctx.Header("Transfer-Encoding", "chunked")

	actorID := ctx.GetUint("userID")
	progressChan := make(chan dto.MigrationProgress)
	errorChan := make(chan error)

	// Jalankan migrasi di Goroutine
	go func() {
		err := c.migrationService.MigrateDataTo(req, actorID, progressChan)
		if err != nil {
			errorChan <- err
		}
		close(progressChan)
		close(errorChan)
	}()

	// Stream data ke client
	ctx.Stream(func(w io.Writer) bool {
		select {
		case progress, ok := <-progressChan:
			if !ok {
				// Selesai
				ctx.SSEvent("complete", map[string]string{"message": "Migrasi Selesai"})
				return false
			}
			// Kirim progress
			ctx.SSEvent("progress", progress)
			return true
		case err := <-errorChan:
			if err != nil {
				ctx.SSEvent("error", map[string]string{"message": err.Error()})
				return false
			}
			return true
		}
	})
}

func (c *ConfigController) RestoreSetup(ctx *gin.Context) {
	isSetup, _ := c.configService.IsSetupComplete()
	if isSetup { APIError(ctx, http.StatusForbidden, "Sudah dikonfigurasi."); return }
	file, err := ctx.FormFile("restore-file")
	if err != nil || !strings.HasSuffix(file.Filename, ".db") { APIError(ctx, http.StatusBadRequest, "File harus .db"); return }
	src, _ := file.Open(); defer src.Close()
	if err := c.backupService.RestoreBackup(src, 0); err != nil { APIError(ctx, http.StatusInternalServerError, "Gagal restore."); return }
	c.configService.SaveConfig(map[string]string{services.IsSetupCompleteKey: "true", "DB_DIALECT": "sqlite", "DB_DSN": "simdokpol.db?_foreign_keys=on"})
	APIResponse(ctx, http.StatusOK, "Restore sukses.", nil)
}

func (c *ConfigController) ShowSetupPage(ctx *gin.Context) {
	if ok, _ := c.configService.IsSetupComplete(); ok { ctx.Redirect(http.StatusFound, "/login"); return }
	cfg, _ := c.configService.GetConfig()
	ctx.HTML(http.StatusOK, "setup.html", gin.H{"Title": "Konfigurasi Awal", "CurrentConfig": cfg})
}