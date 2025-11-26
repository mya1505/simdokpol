package controllers

import (
	"net/http"
	"simdokpol/internal/services"

	"github.com/gin-gonic/gin"
)

type UpdateController struct {
	service        services.UpdateService
	currentVersion string
}

func NewUpdateController(service services.UpdateService, currentVersion string) *UpdateController {
	return &UpdateController{
		service:        service,
		currentVersion: currentVersion,
	}
}

// @Summary Cek Pembaruan Aplikasi
// @Description Mengecek ke GitHub apakah ada versi rilis terbaru.
// @Tags System
// @Produce json
// @Success 200 {object} dto.UpdateCheckResponse
// @Failure 500 {object} map[string]string "Error internal"
// @Security BearerAuth
// @Router /updates/check [get]
func (c *UpdateController) CheckUpdate(ctx *gin.Context) {
	result, err := c.service.CheckForUpdates(c.currentVersion)
	if err != nil {
		// Jangan return 500 agar UI bisa handle graceful degradation
		ctx.JSON(http.StatusOK, gin.H{
			"has_update": false,
			"error":      err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, result)
}