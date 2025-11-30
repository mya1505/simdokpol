package controllers

import (
	"log"
	"net/http"
	"simdokpol/internal/models"
	"simdokpol/internal/services"
	"simdokpol/internal/utils" // Pastikan import utils ada
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type SettingsController struct {
	configService services.ConfigService
	auditService  services.AuditLogService
}

func NewSettingsController(configService services.ConfigService, auditService services.AuditLogService) *SettingsController {
	return &SettingsController{
		configService: configService,
		auditService:  auditService,
	}
}

func (c *SettingsController) GetSettings(ctx *gin.Context) {
	config, err := c.configService.GetConfig()
	if err != nil {
		log.Printf("ERROR: Gagal mengambil data pengaturan: %v", err)
		APIError(ctx, http.StatusInternalServerError, "Gagal mengambil data pengaturan.")
		return
	}
	ctx.JSON(http.StatusOK, config)
}

func (c *SettingsController) UpdateSettings(ctx *gin.Context) {
	var settings map[string]string
	if err := ctx.ShouldBindJSON(&settings); err != nil {
		APIError(ctx, http.StatusBadRequest, "Format data tidak valid")
		return
	}

	// Validasi keamanan Path Traversal
	if path, exists := settings["backup_path"]; exists {
		if strings.Contains(path, "..") {
			APIError(ctx, http.StatusBadRequest, "Path tidak valid.")
			return
		}
	}
	if path, exists := settings["db_dsn"]; exists {
		if strings.Contains(path, "..") {
			APIError(ctx, http.StatusBadRequest, "Path DSN SQLite tidak valid.")
			return
		}
	}

	// Bersihkan password kosong
	if pass, exists := settings["db_pass"]; exists && pass == "" {
		delete(settings, "db_pass")
	}

	// Default SSL Mode
	if ssl, ok := settings["db_sslmode"]; ok && ssl == "" {
		settings["db_sslmode"] = "disable"
	}

	// --- LOGIC DETEKSI RESTART ---
	restartRequired := false
	criticalKeys := []string{"db_dialect", "db_host", "db_port", "db_name", "db_user", "db_pass", "db_dsn", "db_sslmode", "enable_https"}

	for _, key := range criticalKeys {
		if _, exists := settings[key]; exists {
			restartRequired = true
			break
		}
	}

	// --- LOGIC DETEKSI HTTPS AKTIF ---
	// Cek apakah user baru saja mengaktifkan HTTPS (dari sebelumnya mati/tidak ada)
	// Kita kirim flag 'check_https_cert' ke frontend jika enable_https = "true"
	askForCertInstall := false
	if val, ok := settings["enable_https"]; ok && val == "true" {
		askForCertInstall = true
	}

	if err := c.configService.SaveConfig(settings); err != nil {
		log.Printf("ERROR: Gagal menyimpan pengaturan: %v", err)
		APIError(ctx, http.StatusInternalServerError, "Gagal menyimpan pengaturan.")
		return
	}

	actorID := ctx.GetUint("userID")
	logDetail := "Pengaturan sistem telah diperbarui."
	if restartRequired {
		logDetail += " (Restarting System...)"
	}
	c.auditService.LogActivity(actorID, models.AuditSettingsUpdated, logDetail)

	// --- AUTO RESTART SEQUENCE ---
	// Hanya restart otomatis JIKA tidak perlu prompt sertifikat.
	// Jika perlu prompt, frontend yang akan handle restart setelah user klik Yes/No.
	if restartRequired && !askForCertInstall {
		go func() {
			time.Sleep(2 * time.Second)
			log.Println("Melakukan restart otomatis karena perubahan konfigurasi...")
			if err := utils.RestartApp(); err != nil {
				log.Printf("GAGAL RESTART: %v", err)
			}
		}()
	}

	// Kirim respon
	ctx.JSON(http.StatusOK, gin.H{
		"message":          "Pengaturan berhasil disimpan.",
		"restart_required": restartRequired,
		"check_https_cert": askForCertInstall, // <-- Flag Baru
	})
}

// InstallCertificate menghandle request instalasi sertifikat
// @Router /api/settings/install-cert [post]
func (c *SettingsController) InstallCertificate(ctx *gin.Context) {
	// Panggil utilitas sistem
	if err := utils.InstallCertToSystem(); err != nil {
		log.Printf("Gagal install sertifikat: %v", err)
		APIError(ctx, http.StatusInternalServerError, "Gagal menginstal sertifikat. Pastikan Anda klik 'Yes' pada popup Administrator.")
		return
	}
	APIResponse(ctx, http.StatusOK, "Sertifikat berhasil diinstal ke Trusted Root!", nil)
}