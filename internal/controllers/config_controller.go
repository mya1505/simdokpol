package controllers

import (
	"log"
	"net/http"
	"simdokpol/internal/models"
	"simdokpol/internal/services"
	"strings"

	"github.com/gin-gonic/gin"
)

type ConfigController struct {
	configService services.ConfigService
	userService   services.UserService
	backupService services.BackupService
}

func NewConfigController(configService services.ConfigService, userService services.UserService, backupService services.BackupService) *ConfigController {
	return &ConfigController{
		configService: configService,
		userService:   userService,
		backupService: backupService,
	}
}

// SaveSetupRequest mendefinisikan payload yang diterima dari wizard setup
type SaveSetupRequest struct {
	// Langkah 1 & 2: Database
	DBDialect string `json:"db_dialect" binding:"required"`
	DBDSN     string `json:"db_dsn"` // Hanya SQLite
	DBHost    string `json:"db_host"`
	DBPort    string `json:"db_port"`
	DBUser    string `json:"db_user"`
	DBPass    string `json:"db_pass"`
	DBName    string `json:"db_name"`
	DBSSLMode string `json:"db_sslmode"` // <-- FITUR BARU: SSL MODE

	// Langkah 3: Instansi
	KopBaris1   string `json:"kop_baris_1" binding:"required"`
	KopBaris2   string `json:"kop_baris_2" binding:"required"`
	KopBaris3   string `json:"kop_baris_3" binding:"required"`
	NamaKantor  string `json:"nama_kantor" binding:"required"`
	TempatSurat string `json:"tempat_surat" binding:"required"`

	// Langkah 4: Dokumen
	FormatNomorSurat    string `json:"format_nomor_surat" binding:"required"`
	NomorSuratTerakhir  string `json:"nomor_surat_terakhir" binding:"required"`
	ZonaWaktu           string `json:"zona_waktu" binding:"required"`
	ArchiveDurationDays string `json:"archive_duration_days" binding:"required"`

	// Langkah 5: Admin
	AdminNamaLengkap string `json:"admin_nama_lengkap" binding:"required"`
	AdminNRP         string `json:"admin_nrp" binding:"required"`
	AdminPangkat     string `json:"admin_pangkat" binding:"required"`
	AdminPassword    string `json:"admin_password" binding:"required,min=8"`
}

// SaveSetup menangani penyimpanan konfigurasi awal dan pembuatan super admin
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

	// 1. Kumpulkan semua data config dari request ke map
	configData := map[string]string{
		// Database
		"DB_DIALECT": req.DBDialect,
		"DB_DSN":     req.DBDSN,
		"DB_HOST":    req.DBHost,
		"DB_PORT":    req.DBPort,
		"DB_USER":    req.DBUser,
		"DB_PASS":    req.DBPass,
		"DB_NAME":    req.DBName,
		"DB_SSLMODE": req.DBSSLMode, // <-- SIMPAN KE CONFIG & .ENV

		// Instansi
		"kop_baris_1":  req.KopBaris1,
		"kop_baris_2":  req.KopBaris2,
		"kop_baris_3":  req.KopBaris3,
		"nama_kantor":  req.NamaKantor,
		"tempat_surat": req.TempatSurat,

		// Dokumen
		"format_nomor_surat":    req.FormatNomorSurat,
		"nomor_surat_terakhir":  req.NomorSuratTerakhir,
		"zona_waktu":            req.ZonaWaktu,
		"archive_duration_days": req.ArchiveDurationDays,

		// Flag Selesai
		services.IsSetupCompleteKey: "true",
	}

	// Set default DSN jika SQLite dipilih tapi kosong
	if req.DBDialect == "sqlite" && req.DBDSN == "" {
		configData["DB_DSN"] = "simdokpol.db?_foreign_keys=on"
	}

	// 2. Simpan konfigurasi sistem
	if err := c.configService.SaveConfig(configData); err != nil {
		log.Printf("ERROR: Gagal menyimpan konfigurasi sistem saat setup: %v", err)
		APIError(ctx, http.StatusInternalServerError, "Gagal menyimpan konfigurasi sistem.")
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

	// (actorID = 0 menandakan aksi sistem)
	if err := c.userService.Create(superAdmin, 0); err != nil {
		log.Printf("ERROR: Gagal membuat akun Super Admin saat setup: %v", err)
		// Catatan: Config sudah tersimpan, tapi admin gagal.
		// User harus reset manual atau kita handle rollback (kompleks).
		// Untuk saat ini return error cukup.
		APIError(ctx, http.StatusInternalServerError, "Gagal membuat akun Super Admin.")
		return
	}

	APIResponse(ctx, http.StatusOK, "Konfigurasi berhasil disimpan. Silakan login menggunakan akun Super Admin yang baru dibuat.", nil)
}

// RestoreSetup menangani pemulihan database dari file backup saat setup awal
func (c *ConfigController) RestoreSetup(ctx *gin.Context) {
	isSetup, _ := c.configService.IsSetupComplete()
	if isSetup {
		APIError(ctx, http.StatusForbidden, "Aplikasi sudah dikonfigurasi.")
		return
	}

	file, err := ctx.FormFile("restore-file")
	if err != nil {
		APIError(ctx, http.StatusBadRequest, "Tidak ada file yang diunggah.")
		return
	}

	if !strings.HasSuffix(file.Filename, ".db") {
		APIError(ctx, http.StatusBadRequest, "Format file tidak valid. Harap unggah file .db")
		return
	}

	src, err := file.Open()
	if err != nil {
		log.Printf("ERROR: Gagal membuka file restore yang diunggah saat setup: %v", err)
		APIError(ctx, http.StatusInternalServerError, "Gagal memproses file yang diunggah.")
		return
	}
	defer src.Close()

	if err := c.backupService.RestoreBackup(src, 0); err != nil {
		log.Printf("ERROR: Gagal melakukan restore saat setup: %v", err)
		APIError(ctx, http.StatusInternalServerError, "Gagal memulihkan database.")
		return
	}

	configData := map[string]string{
		services.IsSetupCompleteKey: "true",
		"DB_DIALECT":                "sqlite",
		"DB_DSN":                    "simdokpol.db?_foreign_keys=on",
	}
	if err := c.configService.SaveConfig(configData); err != nil {
		log.Printf("ERROR: Gagal menandai setup selesai setelah restore: %v", err)
		APIError(ctx, http.StatusInternalServerError, "Gagal menyimpan status konfigurasi.")
		return
	}

	APIResponse(ctx, http.StatusOK, "Database berhasil dipulihkan. Silakan login.", nil)
}

// ShowSetupPage menampilkan halaman wizard setup
func (c *ConfigController) ShowSetupPage(ctx *gin.Context) {
	isSetup, _ := c.configService.IsSetupComplete()
	if isSetup {
		ctx.Redirect(http.StatusFound, "/login")
		return
	}

	// Ambil config saat ini (dari .env) agar form terisi otomatis jika restart
	currentConfig, _ := c.configService.GetConfig()

	ctx.HTML(http.StatusOK, "setup.html", gin.H{
		"Title":         "Konfigurasi Awal",
		"CurrentConfig": currentConfig,
	})
}