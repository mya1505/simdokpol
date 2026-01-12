package controllers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"simdokpol/internal/dto"
	"simdokpol/internal/models"
	"simdokpol/internal/services"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// DocumentRequest adalah DTO untuk membuat atau memperbarui dokumen.
type DocumentRequest struct {
	NamaLengkap        string `json:"nama_lengkap" binding:"required" example:"BUDI SANTOSO"`
	TempatLahir        string `json:"tempat_lahir" binding:"required" example:"JAKARTA"`
	TanggalLahir       string `json:"tanggal_lahir" binding:"required" example:"1990-01-15"`
	JenisKelamin       string `json:"jenis_kelamin" binding:"required" enums:"Laki-laki,Perempuan"`
	Agama              string `json:"agama" binding:"required" example:"Islam"`
	Pekerjaan          string `json:"pekerjaan" binding:"required" example:"Karyawan Swasta"`
	Alamat             string `json:"alamat" binding:"required" example:"JL. MERDEKA NO. 10, JAKARTA"`
	LokasiHilang       string `json:"lokasi_hilang" binding:"required" example:"Sekitar Pasar Senen"`
	PetugasPelaporID   uint   `json:"petugas_pelapor_id" binding:"required" example:"2"`
	PejabatPersetujuID uint   `json:"pejabat_persetuju_id" binding:"required" example:"1"`
	Items              []struct {
		NamaBarang string `json:"nama_barang" binding:"required" example:"KTP"`
		Deskripsi  string `json:"deskripsi" example:"NIK: 3171234567890001"`
	} `json:"items" binding:"required,min=1"`
}

type LostDocumentController struct {
	docService services.LostDocumentService
}

func NewLostDocumentController(docService services.LostDocumentService) *LostDocumentController {
	return &LostDocumentController{docService: docService}
}

// @Summary Download Dokumen sebagai PDF
// @Router /documents/{id}/pdf [get]
func (c *LostDocumentController) GetPDF(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		APIError(ctx, http.StatusBadRequest, "ID dokumen tidak valid")
		return
	}

	loggedInUserID := ctx.GetUint("userID")

	buffer, filename, err := c.docService.GenerateDocumentPDF(uint(id), loggedInUserID)
	if err != nil {
		if errors.Is(err, services.ErrAccessDenied) {
			APIError(ctx, http.StatusForbidden, "Akses ditolak: Anda tidak memiliki izin untuk melihat dokumen ini.")
			return
		}
		if errors.Is(err, services.ErrNotFound) {
			APIError(ctx, http.StatusNotFound, "Dokumen tidak ditemukan")
			return
		}
		APIError(ctx, http.StatusInternalServerError, "Gagal membuat file PDF.")
		return
	}

	ctx.Header("Content-Description", "File Transfer")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	ctx.Header("Content-Type", "application/pdf")
	ctx.Data(http.StatusOK, "application/pdf", buffer.Bytes())
}

// @Summary Ekspor Dokumen ke Excel
// @Router /documents/export [get]
func (c *LostDocumentController) Export(ctx *gin.Context) {
	query := ctx.Query("q")
	status := ctx.DefaultQuery("status", "active")

	buffer, filename, err := c.docService.ExportDocuments(query, status)
	if err != nil {
		APIError(ctx, http.StatusInternalServerError, "Gagal membuat file ekspor.")
		return
	}

	ctx.Header("Content-Description", "File Transfer")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	ctx.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	ctx.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buffer.Bytes())
}

// @Summary Mendapatkan Dokumen Berdasarkan ID
// @Router /documents/{id} [get]
func (c *LostDocumentController) FindByID(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		APIError(ctx, http.StatusBadRequest, "ID dokumen tidak valid")
		return
	}

	loggedInUserID := ctx.GetUint("userID")

	document, err := c.docService.FindByID(uint(id), loggedInUserID)
	if err != nil {
		if errors.Is(err, services.ErrAccessDenied) {
			APIError(ctx, http.StatusForbidden, "Akses ditolak: Anda tidak memiliki izin untuk melihat dokumen ini.")
			return
		}
		APIError(ctx, http.StatusNotFound, "Dokumen tidak ditemukan")
		return
	}

	ctx.JSON(http.StatusOK, document)
}

// @Summary Pencarian Dokumen Global
// @Router /search [get]
func (c *LostDocumentController) SearchGlobal(ctx *gin.Context) {
	query := ctx.Query("q")

	limitStr := ctx.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	documents, err := c.docService.SearchGlobal(query, limit)
	if err != nil {
		APIError(ctx, http.StatusInternalServerError, "Gagal melakukan pencarian dokumen.")
		return
	}
	ctx.JSON(http.StatusOK, documents)
}

// @Summary Mendapatkan Semua Dokumen (Server-Side Paging)
// @Router /documents [get]
func (c *LostDocumentController) FindAll(ctx *gin.Context) {
	var req dto.DataTableRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		req.Draw = 1
		req.Start = 0
		req.Length = 10
	}
	
	if filter := ctx.Query("filter"); filter != "" {
		req.FilterType = filter
	}

	status := ctx.DefaultQuery("status", "active")

	// --- PERBAIKAN LOGIC: Ambil Context User ---
	userID := ctx.GetUint("userID")
	var userRole string
	if u, exists := ctx.Get("currentUser"); exists {
		if user, ok := u.(*models.User); ok {
			userRole = user.Peran
		}
	}
	// -------------------------------------------

	// Kirim userID dan userRole ke Service
	response, err := c.docService.GetDocumentsPaged(req, status, userID, userRole)
	if err != nil {
		log.Printf("ERROR: Gagal mengambil data dokumen: %v", err)
		ctx.JSON(http.StatusInternalServerError, dto.DataTableResponse{
			Error: "Gagal mengambil data: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// @Summary Menghapus Dokumen
// @Router /documents/{id} [delete]
func (c *LostDocumentController) Delete(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		APIError(ctx, http.StatusBadRequest, "ID dokumen tidak valid")
		return
	}

	loggedInUserID := ctx.GetUint("userID")

	if err := c.docService.DeleteLostDocument(uint(id), loggedInUserID); err != nil {
		if errors.Is(err, services.ErrAccessDenied) {
			APIError(ctx, http.StatusForbidden, err.Error())
			return
		}
		APIError(ctx, http.StatusInternalServerError, "Gagal menghapus dokumen.")
		return
	}

	APIResponse(ctx, http.StatusOK, "Dokumen berhasil dihapus", nil)
}

// @Summary Memperbarui Dokumen
// @Router /documents/{id} [put]
func (c *LostDocumentController) Update(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		APIError(ctx, http.StatusBadRequest, "ID dokumen tidak valid")
		return
	}

	var req DocumentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		APIError(ctx, http.StatusBadRequest, "Input tidak valid: "+err.Error())
		return
	}

	tglLahir, err := time.Parse("2006-01-02", req.TanggalLahir)
	if err != nil {
		APIError(ctx, http.StatusBadRequest, "Format Tanggal Lahir salah, gunakan YYYY-MM-DD")
		return
	}

	loggedInUserID := ctx.GetUint("userID")

	residentData := models.Resident{
		NamaLengkap:  req.NamaLengkap,
		TempatLahir:  req.TempatLahir,
		TanggalLahir: tglLahir,
		JenisKelamin: req.JenisKelamin,
		Agama:        req.Agama,
		Pekerjaan:    req.Pekerjaan,
		Alamat:       req.Alamat,
	}

	var lostItems []models.LostItem
	for _, item := range req.Items {
		lostItems = append(lostItems, models.LostItem{NamaBarang: item.NamaBarang, Deskripsi: item.Deskripsi})
	}

	updatedDoc, err := c.docService.UpdateLostDocument(uint(id), residentData, lostItems, req.LokasiHilang, req.PetugasPelaporID, req.PejabatPersetujuID, loggedInUserID)
	if err != nil {
		if errors.Is(err, services.ErrAccessDenied) {
			APIError(ctx, http.StatusForbidden, err.Error())
			return
		}
		APIError(ctx, http.StatusInternalServerError, "Gagal memperbarui dokumen.")
		return
	}

	ctx.JSON(http.StatusOK, updatedDoc)
}

// @Summary Membuat Dokumen Baru
// @Router /documents [post]
func (c *LostDocumentController) Create(ctx *gin.Context) {
	var req DocumentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		APIError(ctx, http.StatusBadRequest, "Input tidak valid: "+err.Error())
		return
	}

	tglLahir, err := time.Parse("2006-01-02", req.TanggalLahir)
	if err != nil {
		APIError(ctx, http.StatusBadRequest, "Format Tanggal Lahir salah, gunakan YYYY-MM-DD")
		return
	}

	operatorID := ctx.GetUint("userID")

	residentData := models.Resident{
		NamaLengkap:  req.NamaLengkap,
		TempatLahir:  req.TempatLahir,
		TanggalLahir: tglLahir,
		JenisKelamin: req.JenisKelamin,
		Agama:        req.Agama,
		Pekerjaan:    req.Pekerjaan,
		Alamat:       req.Alamat,
	}

	var lostItems []models.LostItem
	for _, item := range req.Items {
		lostItems = append(lostItems, models.LostItem{NamaBarang: item.NamaBarang, Deskripsi: item.Deskripsi})
	}

	createdDoc, err := c.docService.CreateLostDocument(residentData, lostItems, operatorID, req.LokasiHilang, req.PetugasPelaporID, req.PejabatPersetujuID)
	if err != nil {
		APIError(ctx, http.StatusInternalServerError, "Gagal membuat dokumen.")
		return
	}

	ctx.JSON(http.StatusCreated, createdDoc)
}
