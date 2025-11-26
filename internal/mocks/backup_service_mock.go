// internal/mocks/backup_service_mock.go
package mocks

import (
	"io"

	"github.com/stretchr/testify/mock"
)

// BackupService adalah mock untuk services.BackupService
type BackupService struct {
	mock.Mock
}

func (m *BackupService) CreateBackup(actorID uint) (string, error) {
	args := m.Called(actorID)
	return args.String(0), args.Error(1)
}

func (m *BackupService) RestoreBackup(uploadedFile io.Reader, actorID uint) error {
	args := m.Called(uploadedFile, actorID)
	return args.Error(0)
}