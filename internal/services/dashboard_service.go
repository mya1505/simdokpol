package services

import (
	"simdokpol/internal/dto"
	"simdokpol/internal/models"
	"simdokpol/internal/repositories"
	"time"
)

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

// --- PERBAIKAN LOGIC NOTIFIKASI ---
func (s *dashboardService) GetExpiringDocumentsForUser(userID uint, notificationWindowDays int) ([]models.LostDocument, error) {
	appConfig, err := s.configService.GetConfig()
	if err != nil { appConfig = &dto.AppConfig{ArchiveDurationDays: 15} } // Default

	loc, err := s.configService.GetLocation()
	if err != nil { loc = time.UTC }
	
	now := time.Now().In(loc)
	archiveDuration := time.Duration(appConfig.ArchiveDurationDays) * 24 * time.Hour
	notificationWindow := time.Duration(notificationWindowDays) * 24 * time.Hour

	expiryDateStart := now.Add(-archiveDuration)
	expiryDateEnd := expiryDateStart.Add(notificationWindow)

	// 1. Ambil data user yang sedang login
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	// 2. Logic Percabangan Role
	if user.Peran == models.RoleSuperAdmin {
		// Super Admin: Lihat SEMUA notifikasi global
		return s.docRepo.FindAllExpiringDocuments(expiryDateStart, expiryDateEnd)
	} else {
		// Operator: Hanya lihat dokumen miliknya sendiri
		return s.docRepo.FindExpiringDocumentsForUser(userID, expiryDateStart, expiryDateEnd)
	}
}
// ----------------------------------

func (s *dashboardService) GetDashboardStats() (*dto.DashboardStatsDTO, error) {
	loc, err := s.configService.GetLocation()
	if err != nil { loc = time.UTC }
	now := time.Now().In(loc)

	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	endOfDay := startOfDay.AddDate(0, 0, 1).Add(-time.Nanosecond)
	docsToday, _ := s.docRepo.CountByDateRange(startOfDay, endOfDay)

	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Nanosecond)
	docsMonthly, _ := s.docRepo.CountByDateRange(startOfMonth, endOfMonth)

	startOfYear := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, loc)
	endOfYear := startOfYear.AddDate(1, 0, 0).Add(-time.Nanosecond)
	docsYearly, _ := s.docRepo.CountByDateRange(startOfYear, endOfYear)

	activeUsers, _ := s.userRepo.CountAll()

	return &dto.DashboardStatsDTO{
		DocsMonthly: docsMonthly,
		DocsYearly:  docsYearly,
		DocsToday:   docsToday,
		ActiveUsers: activeUsers,
	}, nil
}

func (s *dashboardService) GetMonthlyIssuanceChartData() (*dto.ChartDataDTO, error) {
	loc, err := s.configService.GetLocation()
	if err != nil { loc = time.UTC }
	currentYear := time.Now().In(loc).Year()

	counts, err := s.docRepo.GetMonthlyIssuanceForYear(currentYear)
	if err != nil {
		return &dto.ChartDataDTO{
			Labels: []string{"Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Ags", "Sep", "Okt", "Nov", "Des"},
			Data:   make([]int, 12),
		}, nil
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
		return &dto.PieChartDataDTO{Labels: []string{}, Data: []int{}, BackgroundColors: []string{}}, nil
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

	finalColors := colors
	if len(labels) < len(colors) {
		finalColors = colors[:len(labels)]
	}

	return &dto.PieChartDataDTO{
		Labels:           labels,
		Data:             data,
		BackgroundColors: finalColors,
	}, nil
}