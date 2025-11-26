package controllers

import (
	"fmt" // <-- IMPORT BARU
	"log"
	"net/http"
	"simdokpol/internal/services"

	"github.com/gin-gonic/gin"
)

type AuditLogController struct {
	service services.AuditLogService
}

func NewAuditLogController(service services.AuditLogService) *AuditLogController {
	return &AuditLogController{service: service}
}

// --- HANDLER BARU UNTUK EKSPOR ---
// @Summary Ekspor Log Audit ke Excel
// @Description Mengunduh seluruh riwayat log audit sebagai file .xlsx. Hanya bisa diakses oleh Super Admin.
// @Tags Audit Log
// @Produce application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Success 200 {file} file "File Excel (.xlsx)"
// @Failure 500 {object} map[string]string "Error: Gagal membuat file ekspor"
// @Security BearerAuth
// @Router /audit-logs/export [get]
func (c *AuditLogController) Export(ctx *gin.Context) {
	buffer, filename, err := c.service.ExportAuditLogs()
	if err != nil {
		log.Printf("ERROR: Gagal mengekspor log audit: %v", err)
		APIError(ctx, http.StatusInternalServerError, "Gagal membuat file ekspor.")
		return
	}

	ctx.Header("Content-Description", "File Transfer")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	ctx.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	ctx.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buffer.Bytes())
}
// --- AKHIR HANDLER BARU ---

// @Summary Mendapatkan Semua Log Audit
// @Description Mengambil seluruh riwayat aktivitas yang tercatat di sistem. Hanya bisa diakses oleh Super Admin.
// @Tags Audit Log
// @Produce json
// @Success 200 {array} models.AuditLog
// @Failure 500 {object} map[string]string "Error: Gagal mengambil data log audit"
// @Security BearerAuth
// @Router /audit-logs [get]
// Ganti method FindAll yang lama dengan ini
func (c *AuditLogController) FindAll(ctx *gin.Context) {
    var req dto.DataTableRequest
    // Bind query params dari DataTables (draw, start, length, search)
    if err := ctx.ShouldBindQuery(&req); err != nil {
        // Fallback defaults
        req.Draw = 1
        req.Start = 0
        req.Length = 10
    }

    response, err := c.service.GetAuditLogsPaged(req)
    if err != nil {
        log.Printf("ERROR: Gagal mengambil data audit log: %v", err)
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memuat data"})
        return
    }

    ctx.JSON(http.StatusOK, response)
}