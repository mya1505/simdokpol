package mocks

import (
	"bytes"
	"simdokpol/internal/models"
	"sync" // <-- IMPORT BARU
	"github.com/stretchr/testify/mock"
)

type AuditLogService struct {
	mock.Mock
}

// --- METHOD BARU DITAMBAHKAN ---
func (_m *AuditLogService) SetWaitGroup(wg *sync.WaitGroup) {
	_m.Called(wg)
}
// --- AKHIR METHOD BARU ---

func (_m *AuditLogService) LogActivity(userID uint, action string, details string) {
	_m.Called(userID, action, details)
}

func (_m*AuditLogService) FindAll() ([]models.AuditLog, error) {
	ret := _m.Called()
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).([]models.AuditLog), ret.Error(1)
}

func (_m *AuditLogService) ExportAuditLogs() (*bytes.Buffer, string, error) {
	ret := _m.Called()
	if ret.Get(0) == nil {
		return nil, ret.String(1), ret.Error(2)
	}
	return ret.Get(0).(*bytes.Buffer), ret.String(1), ret.Error(2)
}