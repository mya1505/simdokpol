package repositories

import (
	"fmt"
	"simdokpol/internal/dto"
	"simdokpol/internal/models"
	"time"

	"gorm.io/gorm"
)

type MonthlyCount struct {
	Year  int `gorm:"column:year"`
	Month int `gorm:"column:month"`
	Count int `gorm:"column:count"`
}

type LostDocumentRepository interface {
	Create(tx *gorm.DB, doc *models.LostDocument) (*models.LostDocument, error)
	FindByID(id uint) (*models.LostDocument, error)
	FindAll(query string, statusFilter string, archiveDurationDays int) ([]models.LostDocument, error)
	SearchGlobal(query string) ([]models.LostDocument, error)
	Update(tx *gorm.DB, doc *models.LostDocument) (*models.LostDocument, error)
	Delete(tx *gorm.DB, id uint) error
	GetLastDocumentOfYear(tx *gorm.DB, year int) (*models.LostDocument, error)
	CountByDateRange(start time.Time, end time.Time) (int64, error)
	GetMonthlyIssuanceForYear(year int) ([]MonthlyCount, error)
	GetItemCompositionStats() ([]dto.ItemCompositionStat, error)
	FindExpiringDocumentsForUser(userID uint, expiryDateStart time.Time, expiryDateEnd time.Time) ([]models.LostDocument, error)
	GetItemCompositionStatsInRange(start time.Time, end time.Time) ([]dto.ItemCompositionStat, error)
	CountByOperatorInRange(start time.Time, end time.Time) ([]dto.OperatorStat, error)
	FindAllByDateRange(start time.Time, end time.Time) ([]models.LostDocument, error)
	FindAllPaged(req dto.DataTableRequest, statusFilter string, archiveDurationDays int) ([]models.LostDocument, int64, int64, error)
}

type lostDocumentRepository struct {
	db      *gorm.DB
	dialect string
}

func NewLostDocumentRepository(db *gorm.DB) LostDocumentRepository {
	return &lostDocumentRepository{
		db:      db,
		dialect: db.Dialector.Name(),
	}
}

func (r *lostDocumentRepository) FindAllPaged(req dto.DataTableRequest, statusFilter string, archiveDurationDays int) ([]models.LostDocument, int64, int64, error) {
	var docs []models.LostDocument
	var total int64
	var filtered int64

	// Base Query
	db := r.db.Model(&models.LostDocument{})

	// 1. Hitung Total Data (Sebelum filter search, tapi sesudah filter status)
	archiveDate := time.Now().Add(-time.Duration(archiveDurationDays) * 24 * time.Hour)
	if statusFilter == "archived" {
		db = db.Where("tanggal_laporan <= ?", archiveDate)
	} else {
		// Active
		if r.dialect == "sqlite" {
			db = db.Where("tanggal_laporan > ?", archiveDate)
		} else {
			db = db.Where("tanggal_laporan > ?", archiveDate.Format("2006-01-02 15:04:05"))
		}
	}
	db.Count(&total)

	// 2. Apply Search Filter
	if req.Search != "" {
		searchQuery := fmt.Sprintf("%%%s%%", req.Search)
		// Join diperlukan untuk cari nama resident
		db = db.Joins("JOIN residents ON lost_documents.resident_id = residents.id").
			Where("lost_documents.nomor_surat LIKE ? OR residents.nama_lengkap LIKE ?", searchQuery, searchQuery)
	}
	db.Count(&filtered) // Hitung yang kena filter

	// 3. Apply Paging & Ordering
	// Preload relasi untuk ditampilkan
	db = db.Preload("Resident").
		Preload("Operator").
		Order("lost_documents.tanggal_laporan desc").
		Limit(req.Length).
		Offset(req.Start)

	err := db.Find(&docs).Error
	return docs, total, filtered, err
} 

// --- FUNGSI HELPER UNTUK KUERI DINAMIS ---
func (r *lostDocumentRepository) selectYear(field string) string {
	switch r.dialect {
	case "mysql":
		return fmt.Sprintf("YEAR(%s)", field)
	case "postgres":
		return fmt.Sprintf("EXTRACT(YEAR FROM %s)", field)
	default: // sqlite
		return fmt.Sprintf("CAST(strftime('%%Y', %s) AS INTEGER)", field)
	}
}

func (r *lostDocumentRepository) selectMonth(field string) string {
	switch r.dialect {
	case "mysql":
		return fmt.Sprintf("MONTH(%s)", field)
	case "postgres":
		return fmt.Sprintf("EXTRACT(MONTH FROM %s)", field)
	default: // sqlite
		return fmt.Sprintf("CAST(strftime('%%m', %s) AS INTEGER)", field)
	}
}
// --- AKHIR FUNGSI HELPER ---


func (r *lostDocumentRepository) GetItemCompositionStatsInRange(start time.Time, end time.Time) ([]dto.ItemCompositionStat, error) {
	var results []dto.ItemCompositionStat
	err := r.db.Model(&models.LostItem{}).
		Select("lost_items.nama_barang, COUNT(lost_items.id) as count").
		Joins("JOIN lost_documents ON lost_documents.id = lost_items.lost_document_id").
		Where("lost_documents.tanggal_laporan BETWEEN ? AND ?", start, end).
		Group("lost_items.nama_barang").
		Order("count desc").
		Scan(&results).Error
	return results, err
}

func (r *lostDocumentRepository) CountByOperatorInRange(start time.Time, end time.Time) ([]dto.OperatorStat, error) {
	var results []dto.OperatorStat
	err := r.db.Model(&models.LostDocument{}).
		Select("lost_documents.operator_id, users.nama_lengkap, COUNT(lost_documents.id) as count").
		Joins("JOIN users ON users.id = lost_documents.operator_id").
		Where("lost_documents.tanggal_laporan BETWEEN ? AND ?", start, end).
		Group("lost_documents.operator_id, users.nama_lengkap").
		Order("count desc").
		Scan(&results).Error
	return results, err
}

func (r *lostDocumentRepository) FindAllByDateRange(start time.Time, end time.Time) ([]models.LostDocument, error) {
	var docs []models.LostDocument
	err := r.db.
		Preload("Resident").
		Preload("LostItems").
		Preload("PetugasPelapor").
		Preload("PejabatPersetuju").
		Preload("Operator").
		Preload("LastUpdatedBy").
		Where("tanggal_laporan BETWEEN ? AND ?", start, end).
		Order("tanggal_laporan asc").
		Find(&docs).Error
	return docs, err
}


func (r *lostDocumentRepository) FindExpiringDocumentsForUser(userID uint, expiryDateStart time.Time, expiryDateEnd time.Time) ([]models.LostDocument, error) {
	var docs []models.LostDocument
	err := r.db.
		Where("operator_id = ?", userID).
		Where("tanggal_laporan BETWEEN ? AND ?", expiryDateStart, expiryDateEnd).
		Order("tanggal_laporan asc").
		Find(&docs).Error
	return docs, err
}

func (r *lostDocumentRepository) CountByDateRange(start time.Time, end time.Time) (int64, error) {
	var count int64
	err := r.db.Model(&models.LostDocument{}).Where("tanggal_laporan BETWEEN ? AND ?", start, end).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *lostDocumentRepository) FindAll(query string, statusFilter string, archiveDurationDays int) ([]models.LostDocument, error) {
	var docs []models.LostDocument
	db := r.db.
		Preload("Resident").
		Preload("LostItems").
		Preload("PetugasPelapor").
		Preload("PejabatPersetuju").
		Preload("Operator").
		Order("tanggal_laporan desc")

	archiveDate := time.Now().Add(-time.Duration(archiveDurationDays) * 24 * time.Hour)
	if statusFilter == "archived" {
		db = db.Where("tanggal_laporan <= ?", archiveDate)
	} else {
		if r.dialect == "sqlite" {
			db = db.Where("tanggal_laporan > ?", archiveDate)
		} else {
			db = db.Where("tanggal_laporan > ?", archiveDate.Format("2006-01-02 15:04:05"))
		}
	}

	if query != "" {
		searchQuery := fmt.Sprintf("%%%s%%", query)
		db = db.Joins("JOIN residents ON lost_documents.resident_id = residents.id").
			Where("lost_documents.nomor_surat LIKE ? OR residents.nama_lengkap LIKE ?", searchQuery, searchQuery)
	}

	err := db.Find(&docs).Error
	if err != nil {
		return nil, err
	}
	return docs, nil
}

func (r *lostDocumentRepository) SearchGlobal(query string) ([]models.LostDocument, error) {
	var docs []models.LostDocument
	db := r.db.
		Preload("Resident").
		Preload("LostItems").
		Preload("PetugasPelapor").
		Preload("PejabatPersetuju").
		Preload("Operator").
		Order("tanggal_laporan desc")

	if query != "" {
		searchQuery := fmt.Sprintf("%%%s%%", query)
		db = db.Joins("JOIN residents ON lost_documents.resident_id = residents.id").
			Where("lost_documents.nomor_surat LIKE ? OR residents.nama_lengkap LIKE ?", searchQuery, searchQuery)
	} else {
		return docs, nil
	}

	err := db.Find(&docs).Error
	if err != nil {
		return nil, err
	}
	return docs, nil
}

func (r *lostDocumentRepository) Create(tx *gorm.DB, doc *models.LostDocument) (*models.LostDocument, error) {
	db := r.db
	if tx != nil {
		db = tx
	}
	if err := db.Create(doc).Error; err != nil {
		return nil, err
	}
	return doc, nil
}

func (r *lostDocumentRepository) FindByID(id uint) (*models.LostDocument, error) {
	var doc models.LostDocument
	err := r.db.Preload("Resident").Preload("LostItems").Preload("PetugasPelapor").Preload("PejabatPersetuju").Preload("Operator").Preload("LastUpdatedBy").First(&doc, id).Error
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func (r *lostDocumentRepository) Update(tx *gorm.DB, doc *models.LostDocument) (*models.LostDocument, error) {
	db := r.db
	if tx != nil {
		db = tx
	}
	if err := db.Session(&gorm.Session{FullSaveAssociations: true}).Updates(doc).Error; err != nil {
		return nil, err
	}
	return doc, nil
}

func (r *lostDocumentRepository) Delete(tx *gorm.DB, id uint) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.Delete(&models.LostDocument{}, id).Error
}

func (r *lostDocumentRepository) GetLastDocumentOfYear(tx *gorm.DB, year int) (*models.LostDocument, error) {
	db := r.db
	if tx != nil {
		db = tx
	}
	var doc models.LostDocument
	
	var err error
	switch r.dialect {
	case "mysql":
		err = db.Unscoped().
			Where("YEAR(created_at) = ?", year).
			Order("id desc").
			First(&doc).Error
	case "postgres":
		err = db.Unscoped().
			Where("EXTRACT(YEAR FROM created_at) = ?", year).
			Order("id desc").
			First(&doc).Error
	default: // sqlite
		startOfYear := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
		endOfYear := startOfYear.AddDate(1, 0, 0).Add(-time.Nanosecond)
		err = db.Unscoped().
			Where("created_at BETWEEN ? AND ?", startOfYear, endOfYear).
			Order("id desc").
			First(&doc).Error
	}

	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func (r *lostDocumentRepository) GetMonthlyIssuanceForYear(year int) ([]MonthlyCount, error) {
	var results []MonthlyCount
	
	yearSelect := r.selectYear("tanggal_laporan")
	monthSelect := r.selectMonth("tanggal_laporan")
	
	query := r.db.Model(&models.LostDocument{}).
		Select(fmt.Sprintf("%s as year, %s as month, COUNT(lost_documents.id) as count", yearSelect, monthSelect)).
		Where(fmt.Sprintf("%s = ?", yearSelect), year).
		Group("year, month").
		Order("month asc")
	
	err := query.Scan(&results).Error
	
	return results, err
}

func (r *lostDocumentRepository) GetItemCompositionStats() ([]dto.ItemCompositionStat, error) {
	var results []dto.ItemCompositionStat
	
	// --- PERBAIKAN BUG #1 DI SINI ---
	// Mengganti COUNT(id) menjadi COUNT(lost_items.id) untuk menghindari ambiguitas
	err := r.db.Model(&models.LostItem{}).
		Select("lost_items.nama_barang, COUNT(lost_items.id) as count").
		Joins("JOIN lost_documents ON lost_documents.id = lost_items.lost_document_id").
		Where(fmt.Sprintf("%s = ?", r.selectYear("lost_documents.tanggal_laporan")), time.Now().Year()).
		Group("lost_items.nama_barang").
		Order("count desc").
		Scan(&results).Error
	// --- AKHIR PERBAIKAN ---
		
	return results, err
}