package controllers

import (
	"net/http"
	"simdokpol/internal/dto"
	"simdokpol/internal/services"

	"github.com/gin-gonic/gin"
)

type DBTestController struct {
	service services.DBTestService
}

func NewDBTestController(service services.DBTestService) *DBTestController {
	return &DBTestController{service: service}
}

// @Summary Tes Koneksi Database
// @Description Mencoba koneksi ke database dengan kredensial yang diberikan.
// @Tags Database
// @Accept json
// @Produce json
// @Param credentials body dto.DBTestRequest true "Kredensial Database"
// @Success 200 {object} map[string]string "Pesan Sukses"
// @Failure 400 {object} map[string]string "Error: Input tidak valid"
// @Failure 500 {object} map[string]string "Error: Koneksi Gagal"
// @Security BearerAuth
// @Router /db/test [post]
func (c *DBTestController) TestConnection(ctx *gin.Context) {
	var req dto.DBTestRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		APIError(ctx, http.StatusBadRequest, "Input tidak valid: "+err.Error())
		return
	}

	if err := c.service.TestConnection(req); err != nil {
		// Kirim 500 (Internal Server Error) tapi dengan pesan error yang jelas
		APIError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	APIResponse(ctx, http.StatusOK, "Koneksi berhasil!", nil)
}