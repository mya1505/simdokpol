package utils

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
	"simdokpol/internal/models"
)

const legacyJabatanPrefix = "ANGGOTA JAGA REGU"

func NormalizeLegacyJabatanRegu(db *gorm.DB) error {
	if db == nil {
		return nil
	}

	var users []models.User
	if err := db.Find(&users).Error; err != nil {
		return err
	}

	for _, user := range users {
		jabatan := strings.ToUpper(strings.TrimSpace(user.Jabatan))
		if !strings.HasPrefix(jabatan, legacyJabatanPrefix) {
			continue
		}

		regu := strings.TrimSpace(strings.TrimPrefix(jabatan, legacyJabatanPrefix))
		if regu == "" {
			regu = user.Regu
		}

		user.Jabatan = "ANGGOTA JAGA"
		if user.Regu == "" && regu != "" {
			user.Regu = regu
		}

		if err := db.Save(&user).Error; err != nil {
			return fmt.Errorf("gagal update jabatan user %d: %w", user.ID, err)
		}
	}

	return nil
}
