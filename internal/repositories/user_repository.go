package repositories

import (
	"fmt"
	"simdokpol/internal/dto"
	"simdokpol/internal/models"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *models.User) error
	FindAllPaged(req dto.DataTableRequest, statusFilter string) ([]models.User, int64, int64, error)
	FindByID(id uint) (*models.User, error)
	FindByNRP(nrp string) (*models.User, error)
	FindOperators() ([]models.User, error)
	Update(user *models.User) error
	Delete(id uint) error
	Restore(id uint) error
	CountAll() (int64, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) FindAllPaged(req dto.DataTableRequest, statusFilter string) ([]models.User, int64, int64, error) {
	var users []models.User
	var total, filtered int64

	limit := req.Length
	if limit <= 0 { limit = 10 }
	if limit > 100 { limit = 100 }

	db := r.db.Model(&models.User{})
	if statusFilter == "inactive" {
		db = db.Unscoped().Where("deleted_at IS NOT NULL")
	} else {
		db = db.Where("deleted_at IS NULL")
	}

	db.Count(&total)

	if req.Search != "" {
		search := fmt.Sprintf("%%%s%%", req.Search)
		db = db.Where("nama_lengkap LIKE ? OR nrp LIKE ? OR jabatan LIKE ? OR pangkat LIKE ?", search, search, search, search)
	}
	db.Count(&filtered)

	err := db.Order("nama_lengkap asc").Limit(limit).Offset(req.Start).Find(&users).Error
	return users, total, filtered, err
}

func (r *userRepository) FindByID(id uint) (*models.User, error) {
	var user models.User
	if err := r.db.Unscoped().First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByNRP(nrp string) (*models.User, error) {
	var user models.User
	if err := r.db.Unscoped().Where("nrp = ?", nrp).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// --- PERBAIKAN: KEMBALIKAN FILTER USER AKTIF ---
func (r *userRepository) FindOperators() ([]models.User, error) {
	var users []models.User
	// 1. Hapus Unscoped() agar user yang sudah dihapus tidak muncul.
	// 2. Tetap ambil SEMUA role (Operator & Super Admin) karena Kanit butuh tanda tangan.
	// 3. Filter deleted_at secara eksplisit (optional di GORM, tapi bagus untuk kepastian).
	err := r.db.Where("deleted_at IS NULL").Order("nama_lengkap asc").Find(&users).Error
	return users, err
}

func (r *userRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) Delete(id uint) error {
	return r.db.Delete(&models.User{}, id).Error
}

func (r *userRepository) Restore(id uint) error {
	return r.db.Unscoped().Model(&models.User{}).Where("id = ?", id).Update("deleted_at", nil).Error
}

func (r *userRepository) CountAll() (int64, error) {
	var count int64
	err := r.db.Model(&models.User{}).Count(&count).Error
	return count, err
}