package utils

import (
	"simdokpol/internal/models"

	"gorm.io/gorm"
)

var defaultJobPositions = []string{
	"KAPOLSEK",
	"KANIT SPKT",
	"KASPKT/KASI SPKT",
	"KA SIAGA",
	"PA SIAGA",
	"PETUGAS SPKT",
	"PIKET SPKT",
	"PIKET FUNGSI",
	"OPERATOR PELAYANAN",
	"KANIT JAGA",
	"ANGGOTA JAGA",
}

func EnsureDefaultJobPositions(db *gorm.DB) error {
	if db == nil {
		return nil
	}

	var count int64
	if err := db.Model(&models.JobPosition{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	for _, name := range defaultJobPositions {
		position := models.JobPosition{
			Nama:     name,
			IsActive: true,
		}
		if err := db.Where("nama = ?", name).FirstOrCreate(&position).Error; err != nil {
			return err
		}
	}

	return nil
}
