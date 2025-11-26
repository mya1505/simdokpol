package repositories

import (
	"simdokpol/internal/models"

	"gorm.io/gorm"
)

type ItemTemplateRepository interface {
	Create(template *models.ItemTemplate) error
	FindAll() ([]models.ItemTemplate, error)
	FindAllActive() ([]models.ItemTemplate, error)
	FindByID(id uint) (*models.ItemTemplate, error)
	FindByNamaBarang(nama string) (*models.ItemTemplate, error) // <-- METODE BARU
	Update(template *models.ItemTemplate) error
	Delete(id uint) error
}

type itemTemplateRepository struct {
	db *gorm.DB
}

func NewItemTemplateRepository(db *gorm.DB) ItemTemplateRepository {
	return &itemTemplateRepository{db: db}
}

func (r *itemTemplateRepository) Create(template *models.ItemTemplate) error {
	return r.db.Create(template).Error
}

// FindAll mengambil semua template, termasuk yang di-soft delete (untuk Admin)
func (r *itemTemplateRepository) FindAll() ([]models.ItemTemplate, error) {
	var templates []models.ItemTemplate
	err := r.db.Unscoped().Order("urutan asc, nama_barang asc").Find(&templates).Error
	return templates, err
}

// FindAllActive hanya mengambil template yang aktif (untuk form dokumen)
func (r *itemTemplateRepository) FindAllActive() ([]models.ItemTemplate, error) {
	var templates []models.ItemTemplate
	err := r.db.Where("is_active = ?", true).Order("urutan asc, nama_barang asc").Find(&templates).Error
	return templates, err
}

func (r *itemTemplateRepository) FindByID(id uint) (*models.ItemTemplate, error) {
	var template models.ItemTemplate
	err := r.db.Unscoped().First(&template, id).Error
	return &template, err
}

// --- FUNGSI BARU UNTUK VALIDASI ---
func (r *itemTemplateRepository) FindByNamaBarang(nama string) (*models.ItemTemplate, error) {
	var template models.ItemTemplate
	err := r.db.Where("nama_barang = ? AND is_active = ?", nama, true).First(&template).Error
	return &template, err
}
// --- AKHIR FUNGSI BARU ---

func (r *itemTemplateRepository) Update(template *models.ItemTemplate) error {
	return r.db.Save(template).Error
}

// Delete melakukan soft delete
func (r *itemTemplateRepository) Delete(id uint) error {
	return r.db.Delete(&models.ItemTemplate{}, id).Error
}