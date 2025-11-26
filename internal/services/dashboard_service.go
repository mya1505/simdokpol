package services

import (
	"simdokpol/internal/dto" // <-- IMPORT BARU
	"simdokpol/internal/models"
	"simdokpol/internal/repositories"
	"time"
)

// Struct DTO dipindahkan ke package dto

type DashboardService interface {
	GetDashboardStats() (*dto.DashboardStatsDTO, error)
	GetMonthlyIssuanceChartData() (*dto.ChartDataDTO, error)
	GetItemCompositionPieChartData() (*dto.PieChartDataDTO, error)
	GetExpiringDocumentsForUser(userID uint, notificationWindowDays int) ([]models.LostDocument, error)
}

type dashboardService struct {
	docRepo       repositories.LostDocumentRepository
	userRepo      repositories.UserRepository
	configService ConfigService
}

func NewDashboardService(docRepo repositories.LostDocumentRepository, userRepo repositories.UserRepository, configService ConfigService) DashboardService {
	return &dashboardService{
		docRepo:       docRepo,
		userRepo:      userRepo,
		configService: configService,
	}
}

func (s *dashboardService) GetExpiringDocumentsForUser(userID uint, notificationWindowDays int) ([]models.LostDocument, error) {
	appConfig, err := s.configService.GetConfig()
	if err != nil {
		return nil, err
	}
	loc, err := s.configService.GetLocation()
	if err != nil {
		loc = time.UTC
	}
	now := time.Now().In(loc)

	archiveDuration := time.Duration(appConfig.ArchiveDurationDays) * 24 * time.Hour
	notificationWindow := time.Duration(notificationWindowDays) * 24 * time.Hour

	expiryDateStart := now.Add(-archiveDuration)
	expiryDateEnd := expiryDateStart.Add(notificationWindow)

	return s.docRepo.FindExpiringDocumentsForUser(userID, expiryDateStart, expiryDateEnd)
}

func (s *dashboardService) GetDashboardStats() (*dto.DashboardStatsDTO, error) {
	loc, err := s.configService.GetLocation()
	if err != nil {
		loc = time.UTC
	}
	now := time.Now().In(loc)

	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	endOfDay := startOfDay.AddDate(0, 0, 1).Add(-time.Nanosecond)
	docsToday, err := s.docRepo.CountByDateRange(startOfDay, endOfDay)
	if err != nil {
		return nil, err
	}

	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Nanosecond)
	docsMonthly, err := s.docRepo.CountByDateRange(startOfMonth, endOfMonth)
	if err != nil {
		return nil, err
	}

	startOfYear := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, loc)
	endOfYear := startOfYear.AddDate(1, 0, 0).Add(-time.Nanosecond)
	docsYearly, err := s.docRepo.CountByDateRange(startOfYear, endOfYear)
	if err != nil {
		return nil, err
	}

	activeUsers, err := s.userRepo.CountAll()
	if err != nil {
		return nil, err
	}

	stats := &dto.DashboardStatsDTO{
		DocsMonthly: docsMonthly,
		DocsYearly:  docsYearly,
		DocsToday:   docsToday,
		ActiveUsers: activeUsers,
	}

	return stats, nil
}

func (s *dashboardService) GetMonthlyIssuanceChartData() (*dto.ChartDataDTO, error) {
	loc, err := s.configService.GetLocation()
	if err != nil {
		loc = time.UTC
	}
	currentYear := time.Now().In(loc).Year()

	counts, err := s.docRepo.GetMonthlyIssuanceForYear(currentYear)
	if err != nil {
		return nil, err
	}

	labels := []string{"Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Ags", "Sep", "Okt", "Nov", "Des"}
	data := make([]int, 12)

	for _, count := range counts {
		if count.Month >= 1 && count.Month <= 12 {
			data[count.Month-1] = count.Count
		}
	}

	return &dto.ChartDataDTO{Labels: labels, Data: data}, nil
}

func (s *dashboardService) GetItemCompositionPieChartData() (*dto.PieChartDataDTO, error) {
	stats, err := s.docRepo.GetItemCompositionStats()
	if err != nil {
		return nil, err
	}

	var labels []string
	var data []int
	colors := []string{"#4e73df", "#1cc88a", "#36b9cc", "#f6c23e", "#e74a3b", "#858796"}

	limit := 5
	othersCount := 0
	for i, stat := range stats {
		if i < limit {
			labels = append(labels, stat.NamaBarang)
			data = append(data, stat.Count)
		} else {
			othersCount += stat.Count
		}
	}

	if othersCount > 0 {
		labels = append(labels, "Lainnya")
		data = append(data, othersCount)
	}

	finalColors := colors[:len(labels)]

	return &dto.PieChartDataDTO{
		Labels:           labels,
		Data:             data,
		BackgroundColors: finalColors,
	}, nil
}