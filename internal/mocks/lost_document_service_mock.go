package mocks

import (
	"bytes"
	"simdokpol/internal/dto"
	"simdokpol/internal/models"

	"github.com/stretchr/testify/mock"
)

type LostDocumentService struct {
	mock.Mock
}

func (m *LostDocumentService) CreateLostDocument(residentData models.Resident, items []models.LostItem, operatorID uint, lokasiHilang string, petugasPelaporID uint, pejabatPersetujuID uint) (*models.LostDocument, error) {
	args := m.Called(residentData, items, operatorID, lokasiHilang, petugasPelaporID, pejabatPersetujuID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LostDocument), args.Error(1)
}

func (m *LostDocumentService) UpdateLostDocument(docID uint, residentData models.Resident, items []models.LostItem, lokasiHilang string, petugasPelaporID uint, pejabatPersetujuID uint, loggedInUserID uint) (*models.LostDocument, error) {
	args := m.Called(docID, residentData, items, lokasiHilang, petugasPelaporID, pejabatPersetujuID, loggedInUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LostDocument), args.Error(1)
}

func (m *LostDocumentService) FindAll(query string, statusFilter string) ([]models.LostDocument, error) {
	args := m.Called(query, statusFilter)
	return args.Get(0).([]models.LostDocument), args.Error(1)
}

func (m *LostDocumentService) SearchGlobal(query string, limit int) ([]models.LostDocument, error) {
	args := m.Called(query, limit)
	return args.Get(0).([]models.LostDocument), args.Error(1)
}

func (m *LostDocumentService) FindByID(id uint, actorID uint) (*models.LostDocument, error) {
	args := m.Called(id, actorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LostDocument), args.Error(1)
}

func (m *LostDocumentService) DeleteLostDocument(id uint, loggedInUserID uint) error {
	args := m.Called(id, loggedInUserID)
	return args.Error(0)
}

func (m *LostDocumentService) ExportDocuments(query string, statusFilter string) (*bytes.Buffer, string, error) {
	args := m.Called(query, statusFilter)
	if args.Get(0) == nil {
		return nil, "", args.Error(2)
	}
	return args.Get(0).(*bytes.Buffer), args.String(1), args.Error(2)
}

func (m *LostDocumentService) GenerateDocumentPDF(docID uint, actorID uint) (*bytes.Buffer, string, error) {
	args := m.Called(docID, actorID)
	if args.Get(0) == nil {
		return nil, "", args.Error(2)
	}
	return args.Get(0).(*bytes.Buffer), args.String(1), args.Error(2)
}

// --- FIX DISINI: Update Signature menjadi 4 Parameter ---
func (m *LostDocumentService) GetDocumentsPaged(req dto.DataTableRequest, statusFilter string, userID uint, userRole string) (*dto.DataTableResponse, error) {
	args := m.Called(req, statusFilter, userID, userRole)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.DataTableResponse), args.Error(1)
}
