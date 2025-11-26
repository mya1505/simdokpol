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

func (c *SettingsController) GetSettings(ctx *gin.Context) {
	config, err := c.configService.GetConfig()
	if err != nil {
		log.Printf("ERROR: GetSettings: %v", err)
		APIError(ctx, http.StatusInternalServerError, "Gagal mengambil pengaturan.")
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

	// Security Check: Path Traversal
	if path, ok := settings["backup_path"]; ok && strings.Contains(path, "..") {
		APIError(ctx, http.StatusBadRequest, "Path Backup tidak aman.")
		return
	}
	if path, ok := settings["db_dsn"]; ok && strings.Contains(path, "..") {
		APIError(ctx, http.StatusBadRequest, "Path Database tidak aman.")
		return
	}

	// Bersihkan password kosong (jangan di-overwrite)
	if pass, ok := settings["db_pass"]; ok && pass == "" {
		delete(settings, "db_pass")
	}

	// Default SSL Mode
	if ssl, ok := settings["db_sslmode"]; ok && ssl == "" {
		settings["db_sslmode"] = "disable"
	}

	if err := c.configService.SaveConfig(settings); err != nil {
		log.Printf("ERROR: UpdateSettings: %v", err)
		APIError(ctx, http.StatusInternalServerError, "Gagal menyimpan pengaturan.")
		return
	}

	actorID := ctx.GetUint("userID")
	c.auditService.LogActivity(actorID, models.AuditSettingsUpdated, "Pengaturan sistem diperbarui.")

	APIResponse(ctx, http.StatusOK, "Pengaturan berhasil disimpan.", nil)
}