package controllers

import (
	"net/http"
	"simdokpol/internal/models"
	"simdokpol/internal/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

type JobPositionController struct {
	service services.JobPositionService
}

func NewJobPositionController(service services.JobPositionService) *JobPositionController {
	return &JobPositionController{service: service}
}

type JobPositionRequest struct {
	Nama     string `json:"nama" binding:"required"`
	IsActive *bool  `json:"is_active"`
}

func (c *JobPositionController) FindAll(ctx *gin.Context) {
	data, err := c.service.FindAll()
	if err != nil {
		APIError(ctx, http.StatusInternalServerError, "Gagal mengambil data jabatan.")
		return
	}
	ctx.JSON(http.StatusOK, data)
}

func (c *JobPositionController) FindAllActive(ctx *gin.Context) {
	data, err := c.service.FindAllActive()
	if err != nil {
		APIError(ctx, http.StatusInternalServerError, "Gagal mengambil data jabatan.")
		return
	}
	ctx.JSON(http.StatusOK, data)
}

func (c *JobPositionController) Create(ctx *gin.Context) {
	var req JobPositionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		APIError(ctx, http.StatusBadRequest, "Input tidak valid.")
		return
	}

	position := &models.JobPosition{
		Nama: req.Nama,
	}
	if req.IsActive != nil {
		position.IsActive = *req.IsActive
	}

	actorID := ctx.GetUint("userID")
	if err := c.service.Create(position, actorID); err != nil {
		APIError(ctx, http.StatusBadRequest, err.Error())
		return
	}
	APIResponse(ctx, http.StatusOK, "Jabatan berhasil ditambahkan.", position)
}

func (c *JobPositionController) Update(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		APIError(ctx, http.StatusBadRequest, "ID tidak valid.")
		return
	}

	var req JobPositionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		APIError(ctx, http.StatusBadRequest, "Input tidak valid.")
		return
	}

	position := &models.JobPosition{
		ID:   uint(id),
		Nama: req.Nama,
	}
	if req.IsActive != nil {
		position.IsActive = *req.IsActive
	}

	actorID := ctx.GetUint("userID")
	if err := c.service.Update(position, actorID); err != nil {
		APIError(ctx, http.StatusBadRequest, err.Error())
		return
	}
	APIResponse(ctx, http.StatusOK, "Jabatan berhasil diperbarui.", position)
}

func (c *JobPositionController) Delete(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		APIError(ctx, http.StatusBadRequest, "ID tidak valid.")
		return
	}

	actorID := ctx.GetUint("userID")
	if err := c.service.Delete(uint(id), actorID); err != nil {
		APIError(ctx, http.StatusInternalServerError, "Gagal menonaktifkan jabatan.")
		return
	}
	APIResponse(ctx, http.StatusOK, "Jabatan berhasil dinonaktifkan.", nil)
}

func (c *JobPositionController) Restore(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		APIError(ctx, http.StatusBadRequest, "ID tidak valid.")
		return
	}

	actorID := ctx.GetUint("userID")
	if err := c.service.Restore(uint(id), actorID); err != nil {
		APIError(ctx, http.StatusInternalServerError, "Gagal mengaktifkan jabatan.")
		return
	}
	APIResponse(ctx, http.StatusOK, "Jabatan berhasil diaktifkan.", nil)
}
