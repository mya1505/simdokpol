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
	"sync" // <-- IMPORT BARU
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// setupSettingsTestRouter membuat instance Gin untuk testing SettingsController
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

// Test GET /api/settings
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

// Test PUT /api/settings
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
	
	// --- PERBAIKAN: Set WaitGroup untuk AuditService ---
	var wg sync.WaitGroup
	mockAuditSvc.On("SetWaitGroup", &wg).Once()
	// --- AKHIR PERBAIKAN ---

	mockConfigSvc.On("SaveConfig", mockSettingsUpdate).Return(nil).Once()
	mockAuditSvc.On("LogActivity", adminUserForSettings.ID, models.AuditSettingsUpdated, mock.AnythingOfType("string")).Once()

	jsonBody, _ := json.Marshal(mockSettingsUpdate)
	req, _ := http.NewRequest(http.MethodPut, "/api/settings", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	recorder := httptest.NewRecorder()

	// --- PERBAIKAN: Panggil SetWaitGroup sebelum ServeHTTP ---
	mockAuditSvc.SetWaitGroup(&wg)
	// --- AKHIR PERBAIKAN ---

	router.ServeHTTP(recorder, req)
	
	wg.Wait() // <-- PERBAIKAN: Tunggu goroutine LogActivity selesai

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.JSONEq(t, `{"message":"Pengaturan berhasil disimpan"}`, recorder.Body.String())
	mockConfigSvc.AssertExpectations(t)
	mockAuditSvc.AssertExpectations(t)
}