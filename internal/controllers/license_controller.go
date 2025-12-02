package controllers

import (
	"errors"
	"log"
	"net/http"
	"simdokpol/internal/dto"
	"simdokpol/internal/services"
	"simdokpol/internal/utils" // <-- Tambah Import Utils
	"time"                     // <-- Tambah Import Time

	"github.com/gin-gonic/gin"
)

type LicenseController struct {
	service      services.LicenseService
	auditService services.AuditLogService
}

func (c *LicenseController) GetHardwareID(ctx *gin.Context) {
	hwid := c.service.GetHardwareID()
	ctx.JSON(http.StatusOK, gin.H{"hardware_id": hwid})
}

func NewLicenseController(service services.LicenseService, auditService services.AuditLogService) *LicenseController {
	return &LicenseController{
		service:      service,
		auditService: auditService,
	}
}

// @Summary Aktivasi Lisensi
// @Router /license/activate [post]
func (c *LicenseController) ActivateLicense(ctx *gin.Context) {
	var req dto.LicenseRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		APIError(ctx, http.StatusBadRequest, "Input tidak valid: "+err.Error())
		return
	}

	actorID := ctx.GetUint("userID")

	license, err := c.service.ActivateLicense(req.Key, actorID)
	if err != nil {
		log.Printf("ERROR: Gagal aktivasi lisensi oleh user %d: %v", actorID, err)
		if errors.Is(err, services.ErrLicenseInvalid) || errors.Is(err, services.ErrLicenseBanned) {
			APIError(ctx, http.StatusUnauthorized, err.Error())
			return
		}
		APIError(ctx, http.StatusInternalServerError, "Gagal memproses kunci lisensi.")
		return
	}

	// --- FITUR BARU: AUTO RESTART ---
	go func() {
		time.Sleep(2 * time.Second) // Tunggu response terkirim ke frontend
		log.Println("🔄 Lisensi Aktif. Melakukan Restart Otomatis...")
		utils.RestartApp()
	}()
	// --------------------------------

	APIResponse(ctx, http.StatusOK, "Lisensi berhasil diaktifkan! Sistem akan dimuat ulang...", license)
}