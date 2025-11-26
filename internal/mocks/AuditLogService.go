package mocks

import (
	"bytes"
	"simdokpol/internal/dto" // <-- Pastikan import DTO ada
	"simdokpol/internal/models"
	"sync"

	"github.com/stretchr/testify/mock"
)

type AuditLogService struct {
	mock.Mock
}

func (_m *AuditLogService) SetWaitGroup(wg *sync.WaitGroup) {
	_m.Called(wg)
}

func (_m *AuditLogService) LogActivity(userID uint, action string, details string) {
	_m.Called(userID, action, details)
}

func (_m *AuditLogService) FindAll() ([]models.AuditLog, error) {
	ret := _m.Called()
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).([]models.AuditLog), ret.Error(1)
}

// --- TAMBAHAN BARU (FIX ERROR INTERFACE) ---
func (_m *AuditLogService) GetAuditLogsPaged(req dto.DataTableRequest) (*dto.DataTableResponse, error) {
	ret := _m.Called(req)
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(*dto.DataTableResponse), ret.Error(1)
}
// -------------------------------------------

func (_m *AuditLogService) ExportAuditLogs() (*bytes.Buffer, string, error) {
	ret := _m.Called()
	if ret.Get(0) == nil {
		return nil, ret.String(1), ret.Error(2)
	}
	return ret.Get(0).(*bytes.Buffer), ret.String(1), ret.Error(2)
}