package services

import (
	"simdokpol/internal/models"
	"simdokpol/internal/repositories"
)

type ItemTemplateService interface {
	Create(template *models.ItemTemplate) error
	FindAll() ([]models.ItemTemplate, error)
	FindAllActive() ([]models.ItemTemplate, error)
	FindByID(id uint) (*models.ItemTemplate, error)
	FindByNamaBarang(nama string) (*models.ItemTemplate, error) // <-- METODE BARU
	Update(template *models.ItemTemplate) error
	Delete(id uint) error
}

type itemTemplateService struct {
	repo repositories.ItemTemplateRepository
}

func NewItemTemplateService(repo repositories.ItemTemplateRepository) ItemTemplateService {
	return &itemTemplateService{repo: repo}
}

func (s *itemTemplateService) Create(template *models.ItemTemplate) error {
	// Di masa depan, validasi bisa ditambahkan di sini
	return s.repo.Create(template)
}

func (s *itemTemplateService) FindAll() ([]models.ItemTemplate, error) {
	return s.repo.FindAll()
}

func (s *itemTemplateService) FindAllActive() ([]models.ItemTemplate, error) {
	return s.repo.FindAllActive()
}

func (s *itemTemplateService) FindByID(id uint) (*models.ItemTemplate, error) {
	return s.repo.FindByID(id)
}

// --- FUNGSI BARU UNTUK VALIDASI ---
func (s *itemTemplateService) FindByNamaBarang(nama string) (*models.ItemTemplate, error) {
	return s.repo.FindByNamaBarang(nama)
}
// --- AKHIR FUNGSI BARU ---

func (s *itemTemplateService) Update(template *models.ItemTemplate) error {
	// Di masa depan, validasi bisa ditambahkan di sini
	return s.repo.Update(template)
}

func (s *itemTemplateService) Delete(id uint) error {
	return s.repo.Delete(id)
}