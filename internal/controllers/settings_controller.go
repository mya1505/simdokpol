package controllers

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"simdokpol/internal/models"
	"simdokpol/internal/services"
	"simdokpol/internal/utils"
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

	// Validasi keamanan
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

	// Cek apakah HTTPS baru saja diaktifkan?
	askForCert := false
	if val, ok := settings["enable_https"]; ok && val == "true" {
		askForCert = true
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
	// Restart otomatis hanya jika TIDAK perlu interaksi user (seperti download sertifikat)
	if restartRequired && !askForCert {
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
		"check_https_cert": askForCert,
	})
}

// DownloadCertificate mengizinkan user mengunduh file CRT untuk diinstall manual
// @Router /api/settings/download-cert [get]
func (c *SettingsController) DownloadCertificate(ctx *gin.Context) {
	certDir := filepath.Join(utils.GetAppDataDir(), "certs")
	certPath := filepath.Join(certDir, "ca.crt")

	// Pastikan file ada
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		// Jika belum ada, generate dulu
		utils.EnsureCertificates()
	}

	ctx.Header("Content-Description", "File Transfer")
	ctx.Header("Content-Disposition", "attachment; filename=simdokpol_ca.crt")
	ctx.Header("Content-Type", "application/x-x509-ca-cert")
	ctx.File(certPath)
}

// InstallCertificate mencoba memasang sertifikat ke trust store OS secara otomatis
// @Router /api/settings/install-cert [post]
func (c *SettingsController) InstallCertificate(ctx *gin.Context) {
	certDir := filepath.Join(utils.GetAppDataDir(), "certs")
	certPath := filepath.Join(certDir, "ca.crt")

	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		utils.EnsureCertificates()
	}

	if err := utils.InstallCertificate(certPath); err != nil {
		APIError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Sertifikat berhasil diinstall."})
}
