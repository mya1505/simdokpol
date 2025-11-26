package services

import (
	"bytes"
	"fmt"
	"simdokpol/internal/models"
	"simdokpol/internal/repositories"
	"sync" // <-- IMPORT BARU
	"time"

	"github.com/xuri/excelize/v2"
)

type AuditLogService interface {
	LogActivity(userID uint, action string, details string)
	FindAll() ([]models.AuditLog, error)
	ExportAuditLogs() (*bytes.Buffer, string, error)
	SetWaitGroup(wg *sync.WaitGroup) // <-- METHOD BARU UNTUK TESTING
	GetAuditLogsPaged(req dto.DataTableRequest) (*dto.DataTableResponse, error)
}
}

type auditLogService struct {
	repo repositories.AuditLogRepository
	wg   *sync.WaitGroup // <-- FIELD BARU
}

func NewAuditLogService(repo repositories.AuditLogRepository) AuditLogService {
	return &auditLogService{repo: repo}
}

func (s *auditLogService) GetAuditLogsPaged(req dto.DataTableRequest) (*dto.DataTableResponse, error) {
    logs, total, filtered, err := s.repo.FindAllPaged(req)
    if err != nil {
        return nil, err
    }

    return &dto.DataTableResponse{
        Draw:            req.Draw,
        RecordsTotal:    total,
        RecordsFiltered: filtered,
        Data:            logs,
    }, nil
}

// SetWaitGroup digunakan oleh unit test untuk menyinkronkan goroutine
func (s *auditLogService) SetWaitGroup(wg *sync.WaitGroup) {
	s.wg = wg
}

func (s *auditLogService) ExportAuditLogs() (*bytes.Buffer, string, error) {
	logs, err := s.repo.FindAll()
	if err != nil {
		return nil, "", err
	}

	f := excelize.NewFile()
	sheet := "Data Log Audit"
	if _, err := f.NewSheet(sheet); err != nil {
		return nil, "", err
	}
	f.DeleteSheet("Sheet1")

	headers := []string{"Waktu", "Pengguna (Aktor)", "NRP", "Aksi", "Detail Aktivitas"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, header)
	}

	for i, logEntry := range logs {
		row := i + 2 

		userName := "SISTEM"
		userNRP := "N/A"
		if logEntry.User.ID != 0 {
			userName = logEntry.User.NamaLengkap
			userNRP = logEntry.User.NRP
		}

		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), logEntry.Timestamp.Format("02-01-2006 15:04:05"))
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), userName)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), userNRP)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), logEntry.Aksi)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), logEntry.Detail)
	}

	buffer, err := f.WriteToBuffer()
	if err != nil {
		return nil, "", err
	}

	filename := fmt.Sprintf("Export_Audit_Log_%s.xlsx", time.Now().Format("20060102_150405"))
	return buffer, filename, nil
}

// LogActivity berjalan sebagai goroutine agar tidak memblokir proses utama.
func (s *auditLogService) LogActivity(userID uint, action string, details string) {
	// Jika WaitGroup di-set (hanya saat testing), tambahkan 1
	if s.wg != nil {
		s.wg.Add(1)
	}

	go func() {
		// Jika WaitGroup di-set, panggil Done() saat goroutine selesai
		if s.wg != nil {
			defer s.wg.Done()
		}
		
		logEntry := &models.AuditLog{
			UserID:    userID,
			Aksi:      action,
			Detail:    details,
			Timestamp: time.Now(),
		}
		_ = s.repo.Create(logEntry)
	}()
}

func (s *auditLogService) FindAll() ([]models.AuditLog, error) {
	return s.repo.FindAll()
}