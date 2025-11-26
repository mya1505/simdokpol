package repositories

import (
	"simdokpol/internal/models"
	"gorm.io/gorm"
)

type LicenseRepository interface {
	GetLicense(key string) (*models.License, error)
	SaveLicense(license *models.License) error
}

type licenseRepository struct {
	db *gorm.DB
}

func NewLicenseRepository(db *gorm.DB) LicenseRepository {
	return &licenseRepository{db: db}
}

func (r *licenseRepository) GetLicense(key string) (*models.License, error) {
	var license models.License
	if err := r.db.Preload("ActivatedBy").First(&license, "key = ?", key).Error; err != nil {
		return nil, err
	}
	return &license, nil
}

func (r *licenseRepository) SaveLicense(license *models.License) error {
	return r.db.Save(license).Error
}