package controllers

import (
	"fmt"
	"log"
	"net/http"
	"simdokpol/internal/services"
	"time"

	"github.com/gin-gonic/gin"
)

type ReportController struct {
	reportService services.ReportService
	configService services.ConfigService
}

func NewReportController(reportService services.ReportService, configService services.ConfigService) *ReportController {
	return &ReportController{
		reportService: reportService,
		configService: configService,
	}
}

// ShowReportPage me-render halaman UI untuk memilih rentang tanggal laporan.
// @Tags Reports
// @Security BearerAuth
// @Router /reports/aggregate [get]
func (c *ReportController) ShowReportPage(ctx *gin.Context) {
	user, _ := ctx.Get("currentUser")
	ctx.HTML(http.StatusOK, "report_form.html", gin.H{
		"Title":       "Laporan Agregat",
		"CurrentUser": user,
	})
}

// GenerateReportPDF menangani pembuatan dan pengunduhan laporan PDF.
// @Summary Generate Laporan Agregat PDF
// @Description Membuat laporan PDF agregat berdasarkan rentang tanggal.
// @Tags Reports
// @Produce application/pdf
// @Param start_date query string true "Tanggal Mulai (YYYY-MM-DD)"
// @Param end_date query string true "Tanggal Selesai (YYYY-MM-DD)"
// @Success 200 {file} file "File Laporan PDF"
// @Failure 400 {object} map[string]string "Error: Input tidak valid"
// @Failure 500 {object} map[string]string "Error: Gagal membuat laporan"
// @Security BearerAuth
// @Router /api/reports/aggregate/pdf [get]
func (c *ReportController) GenerateReportPDF(ctx *gin.Context) {
	startDateStr := ctx.Query("start_date")
	endDateStr := ctx.Query("end_date")

	loc, err := c.configService.GetLocation()
	if err != nil {
		loc = time.UTC
	}

	// 1. Validasi Input Tanggal
	start, err := time.ParseInLocation("2006-01-02", startDateStr, loc)
	if err != nil {
		APIError(ctx, http.StatusBadRequest, "Format tanggal mulai tidak valid. Gunakan YYYY-MM-DD.")
		return
	}

	end, err := time.ParseInLocation("2006-01-02", endDateStr, loc)
	if err != nil {
		APIError(ctx, http.StatusBadRequest, "Format tanggal selesai tidak valid. Gunakan YYYY-MM-DD.")
		return
	}
	// Atur 'end' ke akhir hari (23:59:59)
	end = end.Add(24*time.Hour - 1*time.Nanosecond)

	if start.After(end) {
		APIError(ctx, http.StatusBadRequest, "Tanggal mulai tidak boleh lebih besar dari tanggal selesai.")
		return
	}

	// 2. Dapatkan Data dari Service (Tahap 1)
	log.Printf("INFO: Membuat laporan agregat dari %v s/d %v", start, end)
	reportData, err := c.reportService.GenerateAggregateReportData(start, end)
	if err != nil {
		log.Printf("ERROR: Gagal mengambil data laporan agregat: %v", err)
		APIError(ctx, http.StatusInternalServerError, "Gagal mengambil data untuk laporan.")
		return
	}

	// 3. Dapatkan Konfigurasi untuk KOP
	appConfig, err := c.configService.GetConfig()
	if err != nil {
		log.Printf("ERROR: Gagal mengambil konfigurasi untuk PDF: %v", err)
		APIError(ctx, http.StatusInternalServerError, "Gagal memuat konfigurasi aplikasi.")
		return
	}

	// 4. Panggil Generator PDF
	// (Kita akan membuat 'GenerateAggregateReportPDF' di file utils berikutnya)
	buffer, filename, err := c.reportService.GenerateAggregateReportPDF(reportData, appConfig)
	if err != nil {
		log.Printf("ERROR: Gagal men-generate PDF laporan: %v", err)
		APIError(ctx, http.StatusInternalServerError, "Gagal membuat file PDF laporan.")
		return
	}

	// 5. Kirim File ke Klien
	ctx.Header("Content-Description", "File Transfer")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	ctx.Header("Content-Type", "application/pdf")
	ctx.Data(http.StatusOK, "application/pdf", buffer.Bytes())
}