package controllers

import (
	"log"
	"net/http"
	"simdokpol/internal/dto"
	"simdokpol/internal/models"
	"simdokpol/internal/services"
	"strings"

	"github.com/gin-gonic/gin"
)

type ConfigController struct {
	configService    services.ConfigService
	userService      services.UserService
	backupService    services.BackupService
	migrationService services.DataMigrationService
}

// NewConfigController Constructor
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

// SaveSetupRequest definisikan payload setup
type SaveSetupRequest struct {
	// Database
	DBDialect string `json:"db_dialect" binding:"required"`
	DBDSN     string `json:"db_dsn"`
	DBHost    string `json:"db_host"`
	DBPort    string `json:"db_port"`
	DBUser    string `json:"db_user"`
	DBPass    string `json:"db_pass"`
	DBName    string `json:"db_name"`
	DBSSLMode string `json:"db_sslmode"` // Support Postgres SSL

	// Instansi
	KopBaris1   string `json:"kop_baris_1" binding:"required"`
	KopBaris2   string `json:"kop_baris_2" binding:"required"`
	KopBaris3   string `json:"kop_baris_3" binding:"required"`
	NamaKantor  string `json:"nama_kantor" binding:"required"`
	TempatSurat string `json:"tempat_surat" binding:"required"`

	// Dokumen
	FormatNomorSurat    string `json:"format_nomor_surat" binding:"required"`
	NomorSuratTerakhir  string `json:"nomor_surat_terakhir" binding:"required"`
	ZonaWaktu           string `json:"zona_waktu" binding:"required"`
	ArchiveDurationDays string `json:"archive_duration_days" binding:"required"`

	// Admin
	AdminNamaLengkap string `json:"admin_nama_lengkap" binding:"required"`
	AdminNRP         string `json:"admin_nrp" binding:"required"`
	AdminPangkat     string `json:"admin_pangkat" binding:"required"`
	AdminPassword    string `json:"admin_password" binding:"required,min=8"`
}

// SaveSetup menangani simpan konfigurasi awal
func (c *ConfigController) SaveSetup(ctx *gin.Context) {
	isSetup, _ := c.configService.IsSetupComplete()
	if isSetup {
		APIError(ctx, http.StatusForbidden, "Aplikasi sudah dikonfigurasi.")
		return
	}

	var req SaveSetupRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		APIError(ctx, http.StatusBadRequest, "Input tidak valid: "+err.Error())
		return
	}

	// 1. Kumpulkan data ke map
	configData := map[string]string{
		"DB_DIALECT": req.DBDialect,
		"DB_DSN":     req.DBDSN,
		"DB_HOST":    req.DBHost,
		"DB_PORT":    req.DBPort,
		"DB_USER":    req.DBUser,
		"DB_PASS":    req.DBPass,
		"DB_NAME":    req.DBName,
		"DB_SSLMODE": req.DBSSLMode,

		"kop_baris_1":  req.KopBaris1,
		"kop_baris_2":  req.KopBaris2,
		"kop_baris_3":  req.KopBaris3,
		"nama_kantor":  req.NamaKantor,
		"tempat_surat": req.TempatSurat,

		"format_nomor_surat":    req.FormatNomorSurat,
		"nomor_surat_terakhir":  req.NomorSuratTerakhir,
		"zona_waktu":            req.ZonaWaktu,
		"archive_duration_days": req.ArchiveDurationDays,

		services.IsSetupCompleteKey: "true",
	}

	if req.DBDialect == "sqlite" && req.DBDSN == "" {
		configData["DB_DSN"] = "simdokpol.db?_foreign_keys=on"
	}

	// 2. Simpan ke .env dan DB
	if err := c.configService.SaveConfig(configData); err != nil {
		log.Printf("ERROR: Gagal simpan config setup: %v", err)
		APIError(ctx, http.StatusInternalServerError, "Gagal menyimpan konfigurasi.")
		return
	}

	// 3. Buat Super Admin
	superAdmin := &models.User{
		NamaLengkap: req.AdminNamaLengkap,
		NRP:         req.AdminNRP,
		Pangkat:     req.AdminPangkat,
		KataSandi:   req.AdminPassword,
		Peran:       models.RoleSuperAdmin,
		Jabatan:     models.RoleSuperAdmin,
	}

	if err := c.userService.Create(superAdmin, 0); err != nil {
		log.Printf("ERROR: Gagal buat admin setup: %v", err)
		APIError(ctx, http.StatusInternalServerError, "Gagal membuat akun Admin.")
		return
	}

	APIResponse(ctx, http.StatusOK, "Setup selesai. Silakan login.", nil)
}

// MigrateDatabase menangani proses migrasi data antar database
func (c *ConfigController) MigrateDatabase(ctx *gin.Context) {
	// Validasi Input (Target Database Config)
	var req dto.DBTestRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		APIError(ctx, http.StatusBadRequest, "Konfigurasi database tujuan tidak valid.")
		return
	}

	// Validasi Dialect
	if req.DBDialect != "mysql" && req.DBDialect != "postgres" && req.DBDialect != "sqlite" {
		APIError(ctx, http.StatusBadRequest, "Tipe database tidak didukung.")
		return
	}

	actorID := ctx.GetUint("userID")

	// Panggil Service Migrasi
	if err := c.migrationService.MigrateDataTo(req, actorID); err != nil {
		log.Printf("ERROR: Migrasi gagal: %v", err)
		APIError(ctx, http.StatusInternalServerError, "Proses migrasi gagal: "+err.Error())
		return
	}

	APIResponse(ctx, http.StatusOK, "Migrasi data berhasil!", nil)
}

// RestoreSetup menangani restore file .db (SQLite Only) saat setup
func (c *ConfigController) RestoreSetup(ctx *gin.Context) {
	isSetup, _ := c.configService.IsSetupComplete()
	if isSetup {
		APIError(ctx, http.StatusForbidden, "Aplikasi sudah dikonfigurasi.")
		return
	}

	file, err := ctx.FormFile("restore-file")
	if err != nil {
		APIError(ctx, http.StatusBadRequest, "File tidak ditemukan.")
		return
	}

	if !strings.HasSuffix(file.Filename, ".db") {
		APIError(ctx, http.StatusBadRequest, "Format salah. Harus .db")
		return
	}

	src, err := file.Open()
	if err != nil {
		APIError(ctx, http.StatusInternalServerError, "Gagal membuka file.")
		return
	}
	defer src.Close()

	if err := c.backupService.RestoreBackup(src, 0); err != nil {
		log.Printf("ERROR: Restore setup gagal: %v", err)
		APIError(ctx, http.StatusInternalServerError, "Gagal memulihkan database.")
		return
	}

	// Set config jadi sqlite setelah restore
	configData := map[string]string{
		services.IsSetupCompleteKey: "true",
		"DB_DIALECT":                "sqlite",
		"DB_DSN":                    "simdokpol.db?_foreign_keys=on",
	}
	c.configService.SaveConfig(configData)

	APIResponse(ctx, http.StatusOK, "Restore berhasil. Silakan login.", nil)
}

// ShowSetupPage render halaman setup
func (c *ConfigController) ShowSetupPage(ctx *gin.Context) {
	isSetup, _ := c.configService.IsSetupComplete()
	if isSetup {
		ctx.Redirect(http.StatusFound, "/login")
		return
	}

	currentConfig, _ := c.configService.GetConfig()
	ctx.HTML(http.StatusOK, "setup.html", gin.H{
		"Title":         "Konfigurasi Awal",
		"CurrentConfig": currentConfig,
	})
}