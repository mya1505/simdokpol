package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"simdokpol/internal/mocks"
	"simdokpol/internal/models"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Helper untuk setup router
func setupDashboardTestRouter(mockService *mocks.DashboardService, authInjector gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	controller := NewDashboardController(mockService)
	router := gin.New()
	if authInjector != nil {
		router.Use(authInjector)
	}
	// Route sesuai setup di main.go
	router.GET("/api/notifications/expiring-documents", controller.GetExpiringDocuments)
	return router
}

func TestDashboardController_GetExpiringDocuments(t *testing.T) {
	// Data dummy user yang login
	loggedInUser := &models.User{ID: 10, NamaLengkap: "Petugas A"}
	
	// Data dummy dokumen yang akan expired
	now := time.Now()
	mockDocs := []models.LostDocument{
		{ID: 1, NomorSurat: "SKH/001/XI/2025", TanggalLaporan: now},
		{ID: 2, NomorSurat: "SKH/002/XI/2025", TanggalLaporan: now.Add(-24 * time.Hour)},
	}

	t.Run("Sukses - Ada Notifikasi", func(t *testing.T) {
		mockSvc := new(mocks.DashboardService)
		
		// Inject user ke context
		authInjector := func(c *gin.Context) {
			c.Set("currentUser", loggedInUser)
			c.Set("userID", loggedInUser.ID)
			c.Next()
		}

		// Setup mock expectation: 
		// Expect GetExpiringDocumentsForUser dipanggil dengan userID=10 dan window=3 (hardcoded di controller)
		mockSvc.On("GetExpiringDocumentsForUser", loggedInUser.ID, 3).Return(mockDocs, nil).Once()

		router := setupDashboardTestRouter(mockSvc, authInjector)
		req, _ := http.NewRequest("GET", "/api/notifications/expiring-documents", nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)

		// Verifikasi
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response []models.LostDocument
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response, 2) // Harusnya ada 2 dokumen
		assert.Equal(t, "SKH/001/XI/2025", response[0].NomorSurat)
		
		mockSvc.AssertExpectations(t)
	})

	t.Run("Sukses - Tidak Ada Notifikasi", func(t *testing.T) {
		mockSvc := new(mocks.DashboardService)
		authInjector := func(c *gin.Context) {
			c.Set("userID", loggedInUser.ID)
			c.Next()
		}

		// Return list kosong
		mockSvc.On("GetExpiringDocumentsForUser", loggedInUser.ID, 3).Return([]models.LostDocument{}, nil).Once()

		router := setupDashboardTestRouter(mockSvc, authInjector)
		req, _ := http.NewRequest("GET", "/api/notifications/expiring-documents", nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response []models.LostDocument
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Empty(t, response) // Harusnya kosong
	})

	t.Run("Gagal - Service Error (Graceful Degradation)", func(t *testing.T) {
		// Skenario: Jika DB error, frontend tidak boleh error 500, 
		// tapi tetap return 200 OK dengan list kosong (agar lonceng tidak rusak).
		mockSvc := new(mocks.DashboardService)
		authInjector := func(c *gin.Context) {
			c.Set("userID", loggedInUser.ID)
			c.Next()
		}

		// Return error dari service
		mockSvc.On("GetExpiringDocumentsForUser", loggedInUser.ID, 3).Return(nil, errors.New("db error")).Once()

		router := setupDashboardTestRouter(mockSvc, authInjector)
		req, _ := http.NewRequest("GET", "/api/notifications/expiring-documents", nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)

		// Pastikan tetap 200 OK
		assert.Equal(t, http.StatusOK, w.Code)
		
		// Pastikan body-nya array string kosong "[]" (sesuai implementasi controller)
		assert.Equal(t, "[]", w.Body.String()) 
	})
}