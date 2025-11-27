package controllers

import (
	"log"
	"net/http"
	"simdokpol/internal/models"
	"simdokpol/internal/services"
	"simdokpol/internal/utils" // <-- Import Utils
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
	// Cek apakah perubahan memerlukan restart aplikasi
	restartRequired := false
	criticalKeys := []string{"db_dialect", "db_host", "db_port", "db_name", "db_user", "db_pass", "db_dsn", "db_sslmode", "enable_https"}
	
	for _, key := range criticalKeys {
		if _, exists := settings[key]; exists {
			restartRequired = true
			break
		}
	}
	// -----------------------------

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
	if restartRequired {
		go func() {
			// Tunggu 2 detik agar response JSON terkirim ke frontend dulu
			time.Sleep(2 * time.Second)
			log.Println("Melakukan restart otomatis karena perubahan konfigurasi...")
			if err := utils.RestartApp(); err != nil {
				log.Printf("GAGAL RESTART: %v", err)
			}
		}()
	}
	// -----------------------------

	// Kirim respon dengan flag restart_required
	ctx.JSON(http.StatusOK, gin.H{
		"message":          "Pengaturan berhasil disimpan.",
		"restart_required": restartRequired,
	})
}