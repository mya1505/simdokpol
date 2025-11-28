package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"simdokpol/internal/dto"
	"simdokpol/internal/middleware"
	"simdokpol/internal/mocks"
	"simdokpol/internal/models"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupSettingsTestRouter(mockConfigSvc *mocks.ConfigService, mockAuditSvc *mocks.AuditLogService, authInjector gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)

	if mockConfigSvc == nil {
		mockConfigSvc = new(mocks.ConfigService)
	}
	if mockAuditSvc == nil {
		mockAuditSvc = new(mocks.AuditLogService)
	}

	settingsController := NewSettingsController(mockConfigSvc, mockAuditSvc)

	router := gin.New()
	if authInjector != nil {
		router.Use(authInjector)
	}

	adminRoutes := router.Group("/api")
	adminRoutes.Use(middleware.AdminAuthMiddleware())
	{
		adminRoutes.GET("/settings", settingsController.GetSettings)
		adminRoutes.PUT("/settings", settingsController.UpdateSettings)
	}

	return router
}

var adminUserForSettings = &models.User{ID: 1, NamaLengkap: "Admin", Peran: models.RoleSuperAdmin}

func TestSettingsController_GetSettings(t *testing.T) {
	mockConfig := &dto.AppConfig{
		NamaKantor: "POLSEK BAHODOPI",
		KopBaris1:  "KEPOLISIAN NEGARA REPUBLIK INDONESIA",
	}

	mockConfigSvc := new(mocks.ConfigService)
	mockAuditSvc := new(mocks.AuditLogService)
	
	authInjector := func(c *gin.Context) {
		c.Set("currentUser", adminUserForSettings)
		c.Set("userID", adminUserForSettings.ID)
		c.Next()
	}

	router := setupSettingsTestRouter(mockConfigSvc, mockAuditSvc, authInjector)
	
	mockConfigSvc.On("GetConfig").Return(mockConfig, nil).Once()

	req, _ := http.NewRequest(http.MethodGet, "/api/settings", nil)
	req.Header.Set("Accept", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "POLSEK BAHODOPI")
	mockConfigSvc.AssertExpectations(t)
}

func TestSettingsController_UpdateSettings(t *testing.T) {
	mockSettingsUpdate := map[string]string{
		"nama_kantor": "POLSEK BARU",
		"backup_path": "./backups-aman",
	}

	mockConfigSvc := new(mocks.ConfigService)
	mockAuditSvc := new(mocks.AuditLogService)
	
	authInjector := func(c *gin.Context) {
		c.Set("currentUser", adminUserForSettings)
		c.Set("userID", adminUserForSettings.ID)
		c.Next()
	}

	router := setupSettingsTestRouter(mockConfigSvc, mockAuditSvc, authInjector)
	
	var wg sync.WaitGroup
	mockAuditSvc.On("SetWaitGroup", &wg).Once()

	mockConfigSvc.On("SaveConfig", mockSettingsUpdate).Return(nil).Once()
	mockAuditSvc.On("LogActivity", adminUserForSettings.ID, models.AuditSettingsUpdated, mock.AnythingOfType("string")).Once()

	jsonBody, _ := json.Marshal(mockSettingsUpdate)
	req, _ := http.NewRequest(http.MethodPut, "/api/settings", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	recorder := httptest.NewRecorder()

	mockAuditSvc.SetWaitGroup(&wg)

	router.ServeHTTP(recorder, req)
	
	wg.Wait()

	assert.Equal(t, http.StatusOK, recorder.Code)
	
	// --- PERBAIKAN UTAMA: Tambahkan 'restart_required: false' di ekspektasi ---
	// Karena perubahan nama kantor bukan perubahan kritis, restart_required harus false
	expectedJSON := `{"message":"Pengaturan berhasil disimpan.", "restart_required":false}`
	assert.JSONEq(t, expectedJSON, recorder.Body.String())
	// --- AKHIR PERBAIKAN ---

	mockConfigSvc.AssertExpectations(t)
	mockAuditSvc.AssertExpectations(t)
}