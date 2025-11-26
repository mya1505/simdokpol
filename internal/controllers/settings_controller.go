package controllers

import (
	"log"
	"net/http"
	"simdokpol/internal/models"
	"simdokpol/internal/services"
	"strings"

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

// @Summary Mendapatkan Semua Pengaturan Sistem
// @Description Mengambil semua data konfigurasi sistem yang sedang aktif. Hanya bisa diakses oleh Super Admin.
// @Tags Settings
// @Produce json
// @Success 200 {object} dto.AppConfig
// @Failure 500 {object} map[string]string "Error: Gagal mengambil data pengaturan"
// @Security BearerAuth
// @Router /settings [get]
func (c *SettingsController) GetSettings(ctx *gin.Context) {
	config, err := c.configService.GetConfig()
	if err != nil {
		log.Printf("ERROR: Gagal mengambil data pengaturan: %v", err)
		APIError(ctx, http.StatusInternalServerError, "Gagal mengambil data pengaturan.")
		return
	}
	ctx.JSON(http.StatusOK, config)
}

// @Summary Memperbarui Pengaturan Sistem
// @Description Menyimpan satu atau lebih data konfigurasi sistem. Hanya bisa diakses oleh Super Admin.
// @Tags Settings
// @Accept json
// @Produce json
// @Param settings body map[string]string true "Data Pengaturan Baru (semua field AppConfig)"
// @Success 200 {object} map[string]string "Pesan Sukses"
// @Failure 400 {object} map[string]string "Error: Format data tidak valid"
// @Failure 500 {object} map[string]string "Error: Gagal menyimpan pengaturan"
// @Security BearerAuth
// @Router /settings [put]
func (c *SettingsController) UpdateSettings(ctx *gin.Context) {
	var settings map[string]string
	if err := ctx.ShouldBindJSON(&settings); err != nil {
		APIError(ctx, http.StatusBadRequest, "Format data tidak valid")
		return
	}

	// Validasi keamanan sederhana untuk path traversal
	if path, exists := settings["backup_path"]; exists {
		if strings.Contains(path, "..") {
			APIError(ctx, http.StatusBadRequest, "Path Backup tidak valid. Tidak boleh mengandung '..'")
			return
		}
	}

	if path, exists := settings["db_dsn"]; exists {
		if strings.Contains(path, "..") {
			APIError(ctx, http.StatusBadRequest, "Path DSN SQLite tidak valid. Tidak boleh mengandung '..'")
			return
		}
	}

	// Jangan simpan password jika string kosong (artinya user tidak ingin mengubah password)
	if pass, exists := settings["db_pass"]; exists && pass == "" {
		delete(settings, "db_pass")
	}

	// Jangan simpan sslmode jika kosong (default disable)
	if ssl, exists := settings["db_sslmode"]; exists && ssl == "" {
		settings["db_sslmode"] = "disable"
	}

	if err := c.configService.SaveConfig(settings); err != nil {
		log.Printf("ERROR: Gagal menyimpan pengaturan: %v", err)
		APIError(ctx, http.StatusInternalServerError, "Gagal menyimpan pengaturan.")
		return
	}

	actorID := ctx.GetUint("userID")
	c.auditService.LogActivity(actorID, models.AuditSettingsUpdated, "Pengaturan sistem telah diperbarui.")

	APIResponse(ctx, http.StatusOK, "Pengaturan berhasil disimpan", nil)
}