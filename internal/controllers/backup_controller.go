package controllers

import (
	"log"
	"net/http"
	"path/filepath"
	"simdokpol/internal/services"
	"strings"
	// "os" // Uncomment jika ingin auto-delete

	"github.com/gin-gonic/gin"
)

type BackupController struct {
	service services.BackupService
}

func NewBackupController(service services.BackupService) *BackupController {
	return &BackupController{service: service}
}

func (c *BackupController) CreateBackup(ctx *gin.Context) {
	actorID := ctx.GetUint("userID")
	backupPath, err := c.service.CreateBackup(actorID)
	if err != nil {
		log.Printf("ERROR Backup: %v", err)
		APIError(ctx, http.StatusInternalServerError, "Gagal backup.")
		return
	}

	// Kirim file
	fileName := filepath.Base(backupPath)
	ctx.Header("Content-Disposition", "attachment; filename="+fileName)
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.File(backupPath)
	
	// Note: Auto-delete di Windows sulit karena file lock saat dikirim.
	// Biarkan file di folder backups sebagai arsip lokal.
}

func (c *BackupController) RestoreBackup(ctx *gin.Context) {
	file, err := ctx.FormFile("restore-file")
	if err != nil {
		APIError(ctx, http.StatusBadRequest, "File wajib diunggah.")
		return
	}

	if !strings.HasSuffix(file.Filename, ".db") {
		APIError(ctx, http.StatusBadRequest, "Format harus .db")
		return
	}

	src, err := file.Open()
	if err != nil {
		APIError(ctx, http.StatusInternalServerError, "Gagal buka file.")
		return
	}
	defer src.Close()

	actorID := ctx.GetUint("userID")
	if err := c.service.RestoreBackup(src, actorID); err != nil {
		log.Printf("ERROR Restore: %v", err)
		APIError(ctx, http.StatusInternalServerError, "Gagal restore: "+err.Error())
		return
	}

	APIResponse(ctx, http.StatusOK, "Restore berhasil. Silakan restart aplikasi.", nil)
}