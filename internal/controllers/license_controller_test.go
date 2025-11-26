package controllers

import (
	"bytes"
	"encoding/json"
	// "errors" // <-- HAPUS INI (Kita pakai services.ErrLicenseInvalid)
	"net/http"
	"net/http/httptest"
	"simdokpol/internal/middleware"
	"simdokpol/internal/mocks"
	"simdokpol/internal/models"
	"simdokpol/internal/services" // <-- IMPORT SERVICES
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupLicenseTestRouter(mockLicenseSvc *mocks.LicenseService, authInjector gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	
	mockAuditSvc := new(mocks.AuditLogService)
	controller := NewLicenseController(mockLicenseSvc, mockAuditSvc)

	router := gin.New()
	if authInjector != nil {
		router.Use(authInjector)
	}

	api := router.Group("/api")
	api.Use(middleware.AdminAuthMiddleware())
	{
		api.POST("/license/activate", controller.ActivateLicense)
		api.GET("/license/hwid", controller.GetHardwareID)
	}

	return router
}

func TestLicenseController_Activate(t *testing.T) {
	adminUser := &models.User{ID: 1, NamaLengkap: "Admin", Peran: models.RoleSuperAdmin}
	
	t.Run("Sukses Aktivasi", func(t *testing.T) {
		mockSvc := new(mocks.LicenseService)
		authInjector := func(c *gin.Context) {
			c.Set("currentUser", adminUser)
			c.Set("userID", adminUser.ID)
			c.Next()
		}
		
		router := setupLicenseTestRouter(mockSvc, authInjector)
		
		mockLicense := &models.License{Status: "VALID"}
		mockSvc.On("ActivateLicense", "VALID-KEY", adminUser.ID).Return(mockLicense, nil)

		body := map[string]string{"key": "VALID-KEY"}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/api/license/activate", bytes.NewBuffer(jsonBody))
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("Gagal Aktivasi", func(t *testing.T) {
		mockSvc := new(mocks.LicenseService)
		authInjector := func(c *gin.Context) {
			c.Set("currentUser", adminUser)
			c.Set("userID", adminUser.ID)
			c.Next()
		}
		
		router := setupLicenseTestRouter(mockSvc, authInjector)
		
		// --- PERBAIKAN UTAMA DI SINI ---
		// Gunakan services.ErrLicenseInvalid agar errors.Is di controller berhasil mendeteksi jenis errornya
		mockSvc.On("ActivateLicense", "INVALID-KEY", adminUser.ID).Return(nil, services.ErrLicenseInvalid)
		// --- AKHIR PERBAIKAN ---

		body := map[string]string{"key": "INVALID-KEY"}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/api/license/activate", bytes.NewBuffer(jsonBody))
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		// Sekarang kita ekspektasikan 401 (StatusUnauthorized) karena error sudah cocok
		assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	})
}

func TestLicenseController_GetHWID(t *testing.T) {
	mockSvc := new(mocks.LicenseService)
	authInjector := func(c *gin.Context) {
		c.Set("currentUser", &models.User{Peran: models.RoleSuperAdmin})
		c.Next()
	}
	
	router := setupLicenseTestRouter(mockSvc, authInjector)
	
	mockSvc.On("GetHardwareID").Return("TEST-HWID-1234")

	req, _ := http.NewRequest("GET", "/api/license/hwid", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "TEST-HWID-1234")
}