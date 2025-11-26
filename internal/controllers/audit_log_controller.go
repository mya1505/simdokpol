package controllers

import (
	"fmt"
	"log"
	"net/http"
	"simdokpol/internal/dto" // <-- TAMBAH INI (FIX UNDEFINED DTO)
	"simdokpol/internal/services"

	"github.com/gin-gonic/gin"
)

type AuditLogController struct {
	service services.AuditLogService
}

func NewAuditLogController(service services.AuditLogService) *AuditLogController {
	return &AuditLogController{service: service}
}

// @Summary Ekspor Log Audit ke Excel
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

// @Summary Mendapatkan Data Log Audit (Server-Side Paging)
// @Router /audit-logs [get]
func (c *AuditLogController) FindAll(ctx *gin.Context) {
	var req dto.DataTableRequest
	// Bind query params dari DataTables (draw, start, length, search)
	if err := ctx.ShouldBindQuery(&req); err != nil {
		// Fallback defaults jika binding gagal
		req.Draw = 1
		req.Start = 0
		req.Length = 10
	}

	// Panggil service yang sudah support paging
	response, err := c.service.GetAuditLogsPaged(req)
	if err != nil {
		log.Printf("ERROR: Gagal mengambil data audit log: %v", err)
		// Return JSON kosong valid untuk DataTables agar tidak alert error
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"draw": req.Draw, 
			"recordsTotal": 0, 
			"recordsFiltered": 0, 
			"data": []string{}, 
			"error": "Gagal memuat data",
		})
		return
	}

	ctx.JSON(http.StatusOK, response)
}