package repositories

import (
	"simdokpol/internal/models"

	"gorm.io/gorm"
)

type JobPositionRepository interface {
	FindAll() ([]models.JobPosition, error)
	FindAllActive() ([]models.JobPosition, error)
	FindByID(id uint) (*models.JobPosition, error)
	Create(position *models.JobPosition) error
	Update(position *models.JobPosition) error
	Delete(id uint) error
	Restore(id uint) error
}

type jobPositionRepository struct {
	db *gorm.DB
}

func NewJobPositionRepository(db *gorm.DB) JobPositionRepository {
	return &jobPositionRepository{db: db}
}

func (r *jobPositionRepository) FindAll() ([]models.JobPosition, error) {
	var positions []models.JobPosition
	err := r.db.Order("nama asc").Find(&positions).Error
	return positions, err
}

func (r *jobPositionRepository) FindAllActive() ([]models.JobPosition, error) {
	var positions []models.JobPosition
	err := r.db.Where("is_active = ?", true).Order("nama asc").Find(&positions).Error
	return positions, err
}

func (r *jobPositionRepository) FindByID(id uint) (*models.JobPosition, error) {
	var position models.JobPosition
	if err := r.db.Unscoped().First(&position, id).Error; err != nil {
		return nil, err
	}
	return &position, nil
}

func (r *jobPositionRepository) Create(position *models.JobPosition) error {
	return r.db.Create(position).Error
}

func (r *jobPositionRepository) Update(position *models.JobPosition) error {
	return r.db.Save(position).Error
}

func (r *jobPositionRepository) Delete(id uint) error {
	return r.db.Delete(&models.JobPosition{}, id).Error
}

func (r *jobPositionRepository) Restore(id uint) error {
	return r.db.Unscoped().Model(&models.JobPosition{}).Where("id = ?", id).Update("deleted_at", nil).Error
}
