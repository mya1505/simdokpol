package mocks

import (
	"simdokpol/internal/dto" // <-- GANTI IMPORT KE DTO
	"simdokpol/internal/models"
	
	"github.com/stretchr/testify/mock"
)

type DashboardService struct {
	mock.Mock
}

func (m *DashboardService) GetDashboardStats() (*dto.DashboardStatsDTO, error) {
	ret := m.Called()
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(*dto.DashboardStatsDTO), ret.Error(1)
}

func (m *DashboardService) GetMonthlyIssuanceChartData() (*dto.ChartDataDTO, error) {
	ret := m.Called()
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(*dto.ChartDataDTO), ret.Error(1)
}

func (m *DashboardService) GetItemCompositionPieChartData() (*dto.PieChartDataDTO, error) {
	ret := m.Called()
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(*dto.PieChartDataDTO), ret.Error(1)
}

func (m *DashboardService) GetExpiringDocumentsForUser(userID uint, notificationWindowDays int) ([]models.LostDocument, error) {
	ret := m.Called(userID, notificationWindowDays)
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).([]models.LostDocument), ret.Error(1)
}