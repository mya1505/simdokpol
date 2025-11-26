package services

import (
	"bytes"
	"simdokpol/internal/dto" // <-- IMPORT BARU
	// "simdokpol/internal/models" // (Sudah diimpor oleh dto)
	"simdokpol/internal/repositories"
	"simdokpol/internal/utils"
	"time"
)

// --- STRUCT DIHAPUS DARI SINI ---
// AggregateReportData dipindah ke dto
// --- AKHIR PENGHAPUSAN ---

// ReportService mendefinisikan interface untuk layanan laporan.
type ReportService interface {
	GenerateAggregateReportData(start time.Time, end time.Time) (*dto.AggregateReportData, error) // <-- PERUBAHAN TIPE
	GenerateAggregateReportPDF(data *dto.AggregateReportData, config *dto.AppConfig) (*bytes.Buffer, string, error) // <-- PERUBAHAN TIPE
}

type reportService struct {
	docRepo       repositories.LostDocumentRepository
	configService ConfigService
	exeDir        string
}

// NewReportService membuat instance baru dari ReportService.
func NewReportService(docRepo repositories.LostDocumentRepository, configService ConfigService, exeDir string) ReportService {
	return &reportService{
		docRepo:       docRepo,
		configService: configService,
		exeDir:        exeDir,
	}
}

// GenerateAggregateReportData mengambil dan memproses semua data mentah dari database.
func (s *reportService) GenerateAggregateReportData(start time.Time, end time.Time) (*dto.AggregateReportData, error) { // <-- PERUBAHAN TIPE
	totalDocs, err := s.docRepo.CountByDateRange(start, end)
	if err != nil {
		return nil, err
	}

	itemStats, err := s.docRepo.GetItemCompositionStatsInRange(start, end)
	if err != nil {
		return nil, err
	}

	operatorStats, err := s.docRepo.CountByOperatorInRange(start, end)
	if err != nil {
		return nil, err
	}

	docList, err := s.docRepo.FindAllByDateRange(start, end)
	if err != nil {
		return nil, err
	}

	data := &dto.AggregateReportData{ // <-- PERUBAHAN TIPE
		StartDate:       start,
		EndDate:         end,
		TotalDocuments:  totalDocs,
		ItemComposition: itemStats,
		OperatorStats:   operatorStats,
		DocumentList:    docList,
	}

	return data, nil
}

// GenerateAggregateReportPDF memanggil utilitas generator PDF baru.
func (s *reportService) GenerateAggregateReportPDF(data *dto.AggregateReportData, config *dto.AppConfig) (*bytes.Buffer, string, error) { // <-- PERUBAHAN TIPE
	// Panggil utilitas generator PDF yang akan kita buat di tahap berikutnya
	buffer, filename := utils.GenerateAggregateReportPDF(data, config, s.exeDir)
	return buffer, filename, nil
}