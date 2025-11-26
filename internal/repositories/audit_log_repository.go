package repositories

import (
	"fmt"
	"simdokpol/internal/dto" // Import DTO
	"simdokpol/internal/models"

	"gorm.io/gorm"
)

type AuditLogRepository interface {
	Create(log *models.AuditLog) error
	// Ubah FindAll jadi FindAllPaged
	FindAllPaged(req dto.DataTableRequest) ([]models.AuditLog, int64, int64, error)
	FindAll() ([]models.AuditLog, error) // Tetap ada untuk Export Excel
}

type auditLogRepository struct {
	db *gorm.DB
}

func NewAuditLogRepository(db *gorm.DB) AuditLogRepository {
	return &auditLogRepository{db: db}
}

func (r *auditLogRepository) Create(log *models.AuditLog) error {
	return r.db.Create(log).Error
}

func (r *auditLogRepository) FindAll() ([]models.AuditLog, error) {
	var logs []models.AuditLog
	err := r.db.Preload("User").Order("timestamp desc").Find(&logs).Error
	return logs, err
}

// Implementasi Paging Server-Side
func (r *auditLogRepository) FindAllPaged(req dto.DataTableRequest) ([]models.AuditLog, int64, int64, error) {
	var logs []models.AuditLog
	var total int64
	var filtered int64

	db := r.db.Model(&models.AuditLog{})
	
	// 1. Hitung Total
	db.Count(&total)

	// 2. Filter Search
	if req.Search != "" {
		search := fmt.Sprintf("%%%s%%", req.Search)
		db = db.Joins("LEFT JOIN users ON users.id = audit_logs.user_id").
			Where("audit_logs.aksi LIKE ? OR audit_logs.detail LIKE ? OR users.nama_lengkap LIKE ?", search, search, search)
	}
	db.Count(&filtered)

	// 3. Paging & Ordering
	err := db.Preload("User").
		Order("timestamp desc").
		Limit(req.Length).
		Offset(req.Start).
		Find(&logs).Error

	return logs, total, filtered, err
}
