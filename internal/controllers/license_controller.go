package controllers

import (
	"errors"
	"log"
	"net/http"
	"simdokpol/internal/dto"
	"simdokpol/internal/services"

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
// @Description Memvalidasi dan mengaktifkan kunci lisensi (serial key) Pro.
// @Tags License
// @Accept json
// @Produce json
// @Param key body dto.LicenseRequest true "Kunci Lisensi"
// @Success 200 {object} map[string]string "Pesan Sukses"
// @Failure 400 {object} map[string]string "Error: Input tidak valid"
// @Failure 401 {object} map[string]string "Error: Lisensi tidak valid"
// @Failure 500 {object} map[string]string "Error: Terjadi kesalahan pada server"
// @Security BearerAuth
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

	APIResponse(ctx, http.StatusOK, "Lisensi berhasil diaktifkan!", license)
}