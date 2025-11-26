package mocks

import (
	"simdokpol/internal/dto"
	"simdokpol/internal/models"
	"simdokpol/internal/repositories"
	"time"

	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type LostDocumentRepository struct {
	mock.Mock
}

func (_m *LostDocumentRepository) Create(tx *gorm.DB, doc *models.LostDocument) (*models.LostDocument, error) {
	ret := _m.Called(tx, doc)
	return ret.Get(0).(*models.LostDocument), ret.Error(1)
}

func (_m *LostDocumentRepository) FindByID(id uint) (*models.LostDocument, error) {
	ret := _m.Called(id)
	return ret.Get(0).(*models.LostDocument), ret.Error(1)
}

func (_m *LostDocumentRepository) FindAll(query string, statusFilter string, archiveDurationDays int) ([]models.LostDocument, error) {
	ret := _m.Called(query, statusFilter, archiveDurationDays)
	return ret.Get(0).([]models.LostDocument), ret.Error(1)
}

// --- METODE BARU (FIX ERROR) ---
func (_m *LostDocumentRepository) FindAllPaged(req dto.DataTableRequest, statusFilter string, archiveDurationDays int) ([]models.LostDocument, int64, int64, error) {
	ret := _m.Called(req, statusFilter, archiveDurationDays)
	var r0 []models.LostDocument
	if rf, ok := ret.Get(0).(func(dto.DataTableRequest, string, int) []models.LostDocument); ok {
		r0 = rf(req, statusFilter, archiveDurationDays)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]models.LostDocument)
		}
	}
	return r0, ret.Get(1).(int64), ret.Get(2).(int64), ret.Error(3)
}
// ------------------------------

func (_m *LostDocumentRepository) SearchGlobal(query string) ([]models.LostDocument, error) {
	ret := _m.Called(query)
	return ret.Get(0).([]models.LostDocument), ret.Error(1)
}

func (_m *LostDocumentRepository) Update(tx *gorm.DB, doc *models.LostDocument) (*models.LostDocument, error) {
	ret := _m.Called(tx, doc)
	return ret.Get(0).(*models.LostDocument), ret.Error(1)
}

func (_m *LostDocumentRepository) Delete(tx *gorm.DB, id uint) error {
	return _m.Called(tx, id).Error(0)
}

func (_m *LostDocumentRepository) GetLastDocumentOfYear(tx *gorm.DB, year int) (*models.LostDocument, error) {
	ret := _m.Called(tx, year)
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(*models.LostDocument), ret.Error(1)
}

func (_m *LostDocumentRepository) CountByDateRange(start time.Time, end time.Time) (int64, error) {
	ret := _m.Called(start, end)
	return ret.Get(0).(int64), ret.Error(1)
}

func (_m *LostDocumentRepository) GetMonthlyIssuanceForYear(year int) ([]repositories.MonthlyCount, error) {
	ret := _m.Called(year)
	return ret.Get(0).([]repositories.MonthlyCount), ret.Error(1)
}

func (_m *LostDocumentRepository) GetItemCompositionStats() ([]dto.ItemCompositionStat, error) {
	ret := _m.Called()
	return ret.Get(0).([]dto.ItemCompositionStat), ret.Error(1)
}

func (_m *LostDocumentRepository) FindExpiringDocumentsForUser(userID uint, expiryDateStart time.Time, expiryDateEnd time.Time) ([]models.LostDocument, error) {
	ret := _m.Called(userID, expiryDateStart, expiryDateEnd)
	return ret.Get(0).([]models.LostDocument), ret.Error(1)
}

func (_m *LostDocumentRepository) GetItemCompositionStatsInRange(start time.Time, end time.Time) ([]dto.ItemCompositionStat, error) {
	ret := _m.Called(start, end)
	return ret.Get(0).([]dto.ItemCompositionStat), ret.Error(1)
}

func (_m *LostDocumentRepository) CountByOperatorInRange(start time.Time, end time.Time) ([]dto.OperatorStat, error) {
	ret := _m.Called(start, end)
	return ret.Get(0).([]dto.OperatorStat), ret.Error(1)
}

func (_m *LostDocumentRepository) FindAllByDateRange(start time.Time, end time.Time) ([]models.LostDocument, error) {
	ret := _m.Called(start, end)
	return ret.Get(0).([]models.LostDocument), ret.Error(1)
}