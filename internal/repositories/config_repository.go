package repositories

import (
	"simdokpol/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ConfigRepository interface {
	Get(key string) (*models.Configuration, error)
	// --- PERUBAHAN SIGNATURE: Tambahkan *gorm.DB ---
	GetForUpdate(tx *gorm.DB, key string) (*models.Configuration, error)
	GetAll() (map[string]string, error)
	Set(key, value string) error
	// --- PERUBAHAN SIGNATURE: Tambahkan *gorm.DB ---
	SetMultiple(tx *gorm.DB, configs map[string]string) error
}

type configRepository struct {
	db *gorm.DB
}

func NewConfigRepository(db *gorm.DB) ConfigRepository {
	return &configRepository{db: db}
}

// --- PERBAIKAN: Gunakan struct-based Where clause ---
func (r *configRepository) Get(key string) (*models.Configuration, error) {
	var config models.Configuration
	// Menggunakan struct di Where() memaksa GORM untuk meng-quote nama kolom "key"
	if err := r.db.Where(&models.Configuration{Key: key}).First(&config).Error; err != nil {
		return nil, err
	}
	return &config, nil
}
// --- AKHIR PERBAIKAN ---

// --- PERBAIKAN: Gunakan struct-based Where clause ---
func (r *configRepository) GetForUpdate(tx *gorm.DB, key string) (*models.Configuration, error) {
	if tx == nil {
		tx = r.db
	}
	var config models.Configuration
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		// Menggunakan struct di Where() memaksa GORM untuk meng-quote nama kolom "key"
		Where(&models.Configuration{Key: key}).
		First(&config).Error
	
	if err != nil {
		return nil, err
	}
	return &config, nil
}
// --- AKHIR PERBAIKAN ---


func (r *configRepository) GetAll() (map[string]string, error) {
	var configs []models.Configuration
	if err := r.db.Find(&configs).Error; err != nil {
		return nil, err
	}
	configMap := make(map[string]string)
	for _, c := range configs {
		configMap[c.Key] = c.Value
	}
	return configMap, nil
}

func (r *configRepository) Set(key, value string) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value"}),
	}).Create(&models.Configuration{Key: key, Value: value}).Error
}

func (r *configRepository) SetMultiple(tx *gorm.DB, configs map[string]string) error {
	db := tx
	if db == nil {
		db = r.db.Begin()
		defer func() {
			if r := recover(); r != nil {
				db.Rollback()
			}
		}()
	}

	for key, value := range configs {
		config := models.Configuration{Key: key, Value: value}
		if err := db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "key"}},
			DoUpdates: clause.AssignmentColumns([]string{"value"}),
		}).Create(&config).Error; err != nil {
			if tx == nil {
				db.Rollback()
			}
			return err
		}
	}

	if tx == nil {
		return db.Commit().Error
	}
	return nil
}