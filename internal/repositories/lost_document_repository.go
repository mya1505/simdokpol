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
	
	// FindAll Biasa (Untuk Export Excel - Tanpa Paging)
	FindAll(query string, statusFilter string, archiveDurationDays int) ([]models.LostDocument, error)
	
	// FindAllPaged (Untuk DataTables - Dengan Paging & Filter User)
	// Update Signature: Tambah userID dan userRole
	FindAllPaged(req dto.DataTableRequest, statusFilter string, archiveDurationDays int, userID uint, userRole string) ([]models.LostDocument, int64, int64, error)
	
	SearchGlobal(query string, limit int) ([]models.LostDocument, error)
	Update(tx *gorm.DB, doc *models.LostDocument) (*models.LostDocument, error)
	Delete(tx *gorm.DB, id uint) error
	GetLastDocumentOfYear(tx *gorm.DB, year int) (*models.LostDocument, error)
	CountByDateRange(start time.Time, end time.Time) (int64, error)
	GetMonthlyIssuanceForYear(year int) ([]MonthlyCount, error)
	GetItemCompositionStats() ([]dto.ItemCompositionStat, error)
	
	// Method Notifikasi Khusus
	FindExpiringDocumentsForUser(userID uint, expiryDateStart time.Time, expiryDateEnd time.Time) ([]models.LostDocument, error)
	FindAllExpiringDocuments(expiryDateStart time.Time, expiryDateEnd time.Time) ([]models.LostDocument, error)
	
	GetItemCompositionStatsInRange(start time.Time, end time.Time) ([]dto.ItemCompositionStat, error)
	CountByOperatorInRange(start time.Time, end time.Time) ([]dto.OperatorStat, error)
	FindAllByDateRange(start time.Time, end time.Time) ([]models.LostDocument, error)
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

// --- FUNGSI HELPER SQL POLYGLOT ---
func (r *lostDocumentRepository) selectYear(field string) string {
	switch r.dialect {
	case "mysql":
		return fmt.Sprintf("YEAR(%s)", field)
	case "postgres":
		return fmt.Sprintf("CAST(EXTRACT(YEAR FROM %s) AS INTEGER)", field)
	default: // sqlite
		return fmt.Sprintf("CAST(strftime('%%Y', %s) AS INTEGER)", field)
	}
}

func (r *lostDocumentRepository) selectMonth(field string) string {
	switch r.dialect {
	case "mysql":
		return fmt.Sprintf("MONTH(%s)", field)
	case "postgres":
		return fmt.Sprintf("CAST(EXTRACT(MONTH FROM %s) AS INTEGER)", field)
	default: // sqlite
		return fmt.Sprintf("CAST(strftime('%%m', %s) AS INTEGER)", field)
	}
}

// --- IMPLEMENTASI UTAMA ---

func (r *lostDocumentRepository) FindAllPaged(req dto.DataTableRequest, statusFilter string, archiveDurationDays int, userID uint, userRole string) ([]models.LostDocument, int64, int64, error) {
	var docs []models.LostDocument
	var total, filtered int64

	limit := req.Length
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	db := r.db.Model(&models.LostDocument{})

	// 1. Filter Status Dasar (Active/Archived)
	archiveDate := time.Now().Add(-time.Duration(archiveDurationDays) * 24 * time.Hour)
	if statusFilter == "archived" {
		db = db.Where("tanggal_laporan <= ?", archiveDate)
	} else {
		db = db.Where("tanggal_laporan > ?", archiveDate)
	}
	
	// Hitung Total Data (Sesuai Status)
	db.Count(&total)

	// 2. Logic Filter Khusus (Notifikasi Expiring)
	if req.FilterType == "expiring" {
		notificationWindow := 3 
		warningDate := archiveDate.Add(time.Duration(notificationWindow) * 24 * time.Hour)
		
		// Filter rentang waktu expiring
		db = db.Where("tanggal_laporan BETWEEN ? AND ?", archiveDate, warningDate)
		
		// Filter Kepemilikan (Jika bukan Super Admin)
		if userRole != models.RoleSuperAdmin {
			db = db.Where("operator_id = ?", userID)
		}
	}

	// 3. Logic Search (Global Search)
	if req.Search != "" {
		searchQuery := fmt.Sprintf("%%%s%%", req.Search)
		db = db.Joins("JOIN residents ON lost_documents.resident_id = residents.id").
			Where("lost_documents.nomor_surat LIKE ? OR residents.nama_lengkap LIKE ?", searchQuery, searchQuery)
	}
	
	// Hitung Data Terfilter
	db.Count(&filtered)

	// 4. Paging & Ordering
	// Default sort by Tanggal Laporan (ASC untuk expiring biar yg paling mepet di atas, DESC untuk list biasa)
	orderClause := "lost_documents.tanggal_laporan desc"
	if req.FilterType == "expiring" {
		orderClause = "lost_documents.tanggal_laporan asc"
	}

	err := db.Preload("Resident").
		Preload("Operator").
		Preload("PetugasPelapor").
		Preload("PejabatPersetuju").
		Order(orderClause).
		Limit(limit).
		Offset(req.Start).
		Find(&docs).Error

	return docs, total, filtered, err
}

func (r *lostDocumentRepository) FindAll(query string, statusFilter string, archiveDurationDays int) ([]models.LostDocument, error) {
	var docs []models.LostDocument
	db := r.db.Model(&models.LostDocument{})

	archiveDate := time.Now().Add(-time.Duration(archiveDurationDays) * 24 * time.Hour)
	if statusFilter == "archived" {
		db = db.Where("tanggal_laporan <= ?", archiveDate)
	} else {
		db = db.Where("tanggal_laporan > ?", archiveDate)
	}

	if query != "" {
		searchQuery := fmt.Sprintf("%%%s%%", query)
		db = db.Joins("JOIN residents ON lost_documents.resident_id = residents.id").
			Where("lost_documents.nomor_surat LIKE ? OR residents.nama_lengkap LIKE ?", searchQuery, searchQuery)
	}

	err := db.Preload("Resident").
		Preload("LostItems").
		Preload("Operator").
		Preload("PetugasPelapor").
		Preload("PejabatPersetuju").
		Order("lost_documents.tanggal_laporan desc").
		Find(&docs).Error

	return docs, err
}

func (r *lostDocumentRepository) FindByID(id uint) (*models.LostDocument, error) {
	var doc models.LostDocument
	err := r.db.Preload("Resident").Preload("LostItems").Preload("PetugasPelapor").Preload("PejabatPersetuju").Preload("Operator").First(&doc, id).Error
	return &doc, err
}

func (r *lostDocumentRepository) Create(tx *gorm.DB, doc *models.LostDocument) (*models.LostDocument, error) {
	db := r.db
	if tx != nil { db = tx }
	err := db.Create(doc).Error
	return doc, err
}

func (r *lostDocumentRepository) Update(tx *gorm.DB, doc *models.LostDocument) (*models.LostDocument, error) {
	db := r.db
	if tx != nil { db = tx }
	err := db.Session(&gorm.Session{FullSaveAssociations: true}).Updates(doc).Error
	return doc, err
}

func (r *lostDocumentRepository) Delete(tx *gorm.DB, id uint) error {
	db := r.db
	if tx != nil { db = tx }
	return db.Delete(&models.LostDocument{}, id).Error
}

func (r *lostDocumentRepository) SearchGlobal(query string, limit int) ([]models.LostDocument, error) {
	var docs []models.LostDocument

	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	db := r.db.Preload("Resident").Preload("Operator").Order("tanggal_laporan desc")
	if query != "" {
		searchQuery := fmt.Sprintf("%%%s%%", query)
		db = db.Joins("JOIN residents ON lost_documents.resident_id = residents.id").
			Where("lost_documents.nomor_surat LIKE ? OR residents.nama_lengkap LIKE ?", searchQuery, searchQuery)
	}
	err := db.Limit(limit).Find(&docs).Error
	return docs, err
}

func (r *lostDocumentRepository) GetLastDocumentOfYear(tx *gorm.DB, year int) (*models.LostDocument, error) {
	db := r.db
	if tx != nil { db = tx }
	var doc models.LostDocument
	var err error
	yearSel := r.selectYear("created_at")
	err = db.Unscoped().Where(fmt.Sprintf("%s = ?", yearSel), year).Order("id desc").First(&doc).Error
	return &doc, err
}

func (r *lostDocumentRepository) CountByDateRange(start time.Time, end time.Time) (int64, error) {
	var count int64
	err := r.db.Model(&models.LostDocument{}).Where("tanggal_laporan BETWEEN ? AND ?", start, end).Count(&count).Error
	return count, err
}

func (r *lostDocumentRepository) GetMonthlyIssuanceForYear(year int) ([]MonthlyCount, error) {
	var results []MonthlyCount
	yearSel := r.selectYear("tanggal_laporan")
	monthSel := r.selectMonth("tanggal_laporan")
	err := r.db.Model(&models.LostDocument{}).
		Select(fmt.Sprintf("%s as year, %s as month, COUNT(id) as count", yearSel, monthSel)).
		Where(fmt.Sprintf("%s = ?", yearSel), year).
		Group("year, month").
		Order("month asc").
		Scan(&results).Error
	return results, err
}

func (r *lostDocumentRepository) GetItemCompositionStats() ([]dto.ItemCompositionStat, error) {
	var results []dto.ItemCompositionStat
	err := r.db.Model(&models.LostItem{}).
		Select("lost_items.nama_barang, COUNT(lost_items.id) as count").
		Joins("JOIN lost_documents ON lost_documents.id = lost_items.lost_document_id").
		Where(fmt.Sprintf("%s = ?", r.selectYear("lost_documents.tanggal_laporan")), time.Now().Year()).
		Group("lost_items.nama_barang").
		Order("count desc").
		Scan(&results).Error
	return results, err
}

func (r *lostDocumentRepository) FindExpiringDocumentsForUser(userID uint, start time.Time, end time.Time) ([]models.LostDocument, error) {
	var docs []models.LostDocument
	err := r.db.Where("operator_id = ? AND tanggal_laporan BETWEEN ? AND ?", userID, start, end).Find(&docs).Error
	return docs, err
}

func (r *lostDocumentRepository) FindAllExpiringDocuments(start time.Time, end time.Time) ([]models.LostDocument, error) {
	var docs []models.LostDocument
	err := r.db.Where("tanggal_laporan BETWEEN ? AND ?", start, end).Find(&docs).Error
	return docs, err
}

func (r *lostDocumentRepository) GetItemCompositionStatsInRange(start time.Time, end time.Time) ([]dto.ItemCompositionStat, error) {
	var results []dto.ItemCompositionStat
	err := r.db.Model(&models.LostItem{}).
		Select("lost_items.nama_barang, COUNT(lost_items.id) as count").
		Joins("JOIN lost_documents ON lost_documents.id = lost_items.lost_document_id").
		Where("lost_documents.tanggal_laporan BETWEEN ? AND ?", start, end).
		Group("lost_items.nama_barang").Order("count desc").Scan(&results).Error
	return results, err
}

func (r *lostDocumentRepository) CountByOperatorInRange(start time.Time, end time.Time) ([]dto.OperatorStat, error) {
	var results []dto.OperatorStat
	err := r.db.Model(&models.LostDocument{}).
		Select("lost_documents.operator_id, users.nama_lengkap, COUNT(lost_documents.id) as count").
		Joins("JOIN users ON users.id = lost_documents.operator_id").
		Where("lost_documents.tanggal_laporan BETWEEN ? AND ?", start, end).
		Group("lost_documents.operator_id, users.nama_lengkap").Order("count desc").Scan(&results).Error
	return results, err
}

func (r *lostDocumentRepository) FindAllByDateRange(start time.Time, end time.Time) ([]models.LostDocument, error) {
	var docs []models.LostDocument
	err := r.db.Preload("Resident").Preload("Operator").Preload("LostItems").Preload("PetugasPelapor").Preload("PejabatPersetuju").
		Where("tanggal_laporan BETWEEN ? AND ?", start, end).Order("tanggal_laporan asc").Find(&docs).Error
	return docs, err
}
