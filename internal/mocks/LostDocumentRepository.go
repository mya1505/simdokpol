package mocks

import (
	"simdokpol/internal/dto"
	"simdokpol/internal/models"
	"simdokpol/internal/repositories" // Import interface asli untuk tipe MonthlyCount
	"time"

	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type LostDocumentRepository struct {
	mock.Mock
}

func (m *LostDocumentRepository) Create(tx *gorm.DB, doc *models.LostDocument) (*models.LostDocument, error) {
	args := m.Called(tx, doc)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LostDocument), args.Error(1)
}

func (m *LostDocumentRepository) FindByID(id uint) (*models.LostDocument, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LostDocument), args.Error(1)
}

func (m *LostDocumentRepository) FindAll(query string, statusFilter string, archiveDurationDays int) ([]models.LostDocument, error) {
	args := m.Called(query, statusFilter, archiveDurationDays)
	return args.Get(0).([]models.LostDocument), args.Error(1)
}

// FIX: Update Signature (Tambah userID & userRole)
func (m *LostDocumentRepository) FindAllPaged(req dto.DataTableRequest, statusFilter string, archiveDurationDays int, userID uint, userRole string) ([]models.LostDocument, int64, int64, error) {
	args := m.Called(req, statusFilter, archiveDurationDays, userID, userRole)
	return args.Get(0).([]models.LostDocument), args.Get(1).(int64), args.Get(2).(int64), args.Error(3)
}

func (m *LostDocumentRepository) SearchGlobal(query string, limit int) ([]models.LostDocument, error) {
	args := m.Called(query, limit)
	return args.Get(0).([]models.LostDocument), args.Error(1)
}

func (m *LostDocumentRepository) Update(tx *gorm.DB, doc *models.LostDocument) (*models.LostDocument, error) {
	args := m.Called(tx, doc)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LostDocument), args.Error(1)
}

func (m *LostDocumentRepository) Delete(tx *gorm.DB, id uint) error {
	args := m.Called(tx, id)
	return args.Error(0)
}

func (m *LostDocumentRepository) GetLastDocumentOfYear(tx *gorm.DB, year int) (*models.LostDocument, error) {
	args := m.Called(tx, year)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LostDocument), args.Error(1)
}

func (m *LostDocumentRepository) CountByDateRange(start time.Time, end time.Time) (int64, error) {
	args := m.Called(start, end)
	return args.Get(0).(int64), args.Error(1)
}

func (m *LostDocumentRepository) GetMonthlyIssuanceForYear(year int) ([]repositories.MonthlyCount, error) {
	args := m.Called(year)
	return args.Get(0).([]repositories.MonthlyCount), args.Error(1)
}

func (m *LostDocumentRepository) GetItemCompositionStats() ([]dto.ItemCompositionStat, error) {
	args := m.Called()
	return args.Get(0).([]dto.ItemCompositionStat), args.Error(1)
}

// --- FIX: TAMBAH METHOD BARU UNTUK NOTIFIKASI ---
func (m *LostDocumentRepository) FindExpiringDocumentsForUser(userID uint, start time.Time, end time.Time) ([]models.LostDocument, error) {
	args := m.Called(userID, start, end)
	return args.Get(0).([]models.LostDocument), args.Error(1)
}

func (m *LostDocumentRepository) FindAllExpiringDocuments(start time.Time, end time.Time) ([]models.LostDocument, error) {
	args := m.Called(start, end)
	return args.Get(0).([]models.LostDocument), args.Error(1)
}
// ------------------------------------------------

func (m *LostDocumentRepository) GetItemCompositionStatsInRange(start time.Time, end time.Time) ([]dto.ItemCompositionStat, error) {
	args := m.Called(start, end)
	return args.Get(0).([]dto.ItemCompositionStat), args.Error(1)
}

func (m *LostDocumentRepository) CountByOperatorInRange(start time.Time, end time.Time) ([]dto.OperatorStat, error) {
	args := m.Called(start, end)
	return args.Get(0).([]dto.OperatorStat), args.Error(1)
}

func (m *LostDocumentRepository) FindAllByDateRange(start time.Time, end time.Time) ([]models.LostDocument, error) {
	args := m.Called(start, end)
	return args.Get(0).([]models.LostDocument), args.Error(1)
}
