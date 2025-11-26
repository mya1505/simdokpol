package controllers

import (
	"net/http"
	"simdokpol/internal/models"
	"simdokpol/internal/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ItemTemplateController struct {
	service services.ItemTemplateService
}

func NewItemTemplateController(service services.ItemTemplateService) *ItemTemplateController {
	return &ItemTemplateController{service: service}
}

// @Summary Mendapatkan Semua Template Barang (Aktif)
// @Description Mengambil daftar template barang yang aktif untuk form dokumen.
// @Tags Item Templates
// @Produce json
// @Success 200 {array} models.ItemTemplate
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /api/item-templates/active [get]
func (c *ItemTemplateController) FindAllActive(ctx *gin.Context) {
	templates, err := c.service.FindAllActive()
	if err != nil {
		APIError(ctx, http.StatusInternalServerError, "Gagal mengambil data template")
		return
	}
	ctx.JSON(http.StatusOK, templates)
}

// @Summary Mendapatkan Semua Template Barang (Admin)
// @Description Mengambil semua template barang, termasuk yang non-aktif (Super Admin).
// @Tags Item Templates
// @Produce json
// @Success 200 {array} models.ItemTemplate
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /api/item-templates [get]
func (c *ItemTemplateController) FindAll(ctx *gin.Context) {
	templates, err := c.service.FindAll()
	if err != nil {
		APIError(ctx, http.StatusInternalServerError, "Gagal mengambil data template")
		return
	}
	ctx.JSON(http.StatusOK, templates)
}

// @Summary Mendapatkan Template Barang per ID (Admin)
// @Description Mengambil detail satu template barang (Super Admin).
// @Tags Item Templates
// @Produce json
// @Param id path int true "ID Template"
// @Success 200 {object} models.ItemTemplate
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /api/item-templates/{id} [get]
func (c *ItemTemplateController) FindByID(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		APIError(ctx, http.StatusBadRequest, "ID template tidak valid")
		return
	}
	template, err := c.service.FindByID(uint(id))
	if err != nil {
		APIError(ctx, http.StatusNotFound, "Template tidak ditemukan")
		return
	}
	ctx.JSON(http.StatusOK, template)
}

// @Summary Membuat Template Barang (Admin)
// @Description Membuat template barang baru (Super Admin).
// @Tags Item Templates
// @Accept json
// @Produce json
// @Param template body models.ItemTemplate true "Data Template Baru"
// @Success 201 {object} models.ItemTemplate
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /api/item-templates [post]
func (c *ItemTemplateController) Create(ctx *gin.Context) {
	var template models.ItemTemplate
	if err := ctx.ShouldBindJSON(&template); err != nil {
		APIError(ctx, http.StatusBadRequest, "Input tidak valid: "+err.Error())
		return
	}

	if err := c.service.Create(&template); err != nil {
		APIError(ctx, http.StatusInternalServerError, "Gagal menyimpan template")
		return
	}
	ctx.JSON(http.StatusCreated, template)
}

// @Summary Memperbarui Template Barang (Admin)
// @Description Memperbarui template barang yang ada (Super Admin).
// @Tags Item Templates
// @Accept json
// @Produce json
// @Param id path int true "ID Template"
// @Param template body models.ItemTemplate true "Data Template"
// @Success 200 {object} models.ItemTemplate
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /api/item-templates/{id} [put]
func (c *ItemTemplateController) Update(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		APIError(ctx, http.StatusBadRequest, "ID template tidak valid")
		return
	}

	var template models.ItemTemplate
	if err := ctx.ShouldBindJSON(&template); err != nil {
		APIError(ctx, http.StatusBadRequest, "Input tidak valid: "+err.Error())
		return
	}
	template.ID = uint(id) // Pastikan ID-nya benar

	if err := c.service.Update(&template); err != nil {
		APIError(ctx, http.StatusInternalServerError, "Gagal memperbarui template")
		return
	}
	ctx.JSON(http.StatusOK, template)
}

// @Summary Menghapus Template Barang (Admin)
// @Description Soft delete template barang (Super Admin).
// @Tags Item Templates
// @Produce json
// @Param id path int true "ID Template"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /api/item-templates/{id} [delete]
func (c *ItemTemplateController) Delete(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		APIError(ctx, http.StatusBadRequest, "ID template tidak valid")
		return
	}

	if err := c.service.Delete(uint(id)); err != nil {
		APIError(ctx, http.StatusInternalServerError, "Gagal menghapus template")
		return
	}
	APIResponse(ctx, http.StatusOK, "Template berhasil dihapus (soft delete).", nil)
}