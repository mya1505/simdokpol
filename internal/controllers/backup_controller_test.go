package controllers

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"simdokpol/internal/middleware"
	"simdokpol/internal/mocks"
	"simdokpol/internal/models"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// setupBackupTestRouter membuat instance Gin untuk testing BackupController
func setupBackupTestRouter(mockBackupSvc *mocks.BackupService, authInjector gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	if mockBackupSvc == nil {
		mockBackupSvc = new(mocks.BackupService)
	}
	backupController := NewBackupController(mockBackupSvc)
	router := gin.New()
	if authInjector != nil {
		router.Use(authInjector)
	}
	adminRoutes := router.Group("/api")
	adminRoutes.Use(middleware.AdminAuthMiddleware())
	{
		adminRoutes.POST("/backups", backupController.CreateBackup)
		adminRoutes.POST("/restore", backupController.RestoreBackup)
	}
	return router
}

var adminUserForBackup = &models.User{ID: 1, NamaLengkap: "Admin", Peran: models.RoleSuperAdmin}

// Test POST /api/backups
func TestBackupController_CreateBackup(t *testing.T) {
	mockBackupSvc := new(mocks.BackupService)
	authInjector := func(c *gin.Context) {
		c.Set("currentUser", adminUserForBackup)
		c.Set("userID", adminUserForBackup.ID)
		c.Next()
	}
	router := setupBackupTestRouter(mockBackupSvc, authInjector)

	// Buat dummy file di direktori temp yang akan bersih otomatis
	tempDir := t.TempDir()
	dummyFileName := "test_backup.db"
	dummyFilePath := filepath.Join(tempDir, dummyFileName)
	err := os.WriteFile(dummyFilePath, []byte("dummy db data"), 0644)
	assert.NoError(t, err)
	
	// Mock service untuk MENGEMBALIKAN path ke file dummy itu
	mockBackupSvc.On("CreateBackup", adminUserForBackup.ID).Return(dummyFilePath, nil).Once()

	req, _ := http.NewRequest(http.MethodPost, "/api/backups", nil)
	req.Header.Set("Accept", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	
	// --- PERBAIKAN: Menambahkan tanda kutip di sekitar nama file ---
	expectedHeader := `attachment; filename="` + dummyFileName + `"`
	assert.Contains(t, recorder.Header().Get("Content-Disposition"), expectedHeader)
	// --- AKHIR PERBAIKAN ---

	assert.Equal(t, "dummy db data", recorder.Body.String())
	mockBackupSvc.AssertExpectations(t)
}

// Test POST /api/restore
func TestBackupController_RestoreBackup(t *testing.T) {
	mockBackupSvc := new(mocks.BackupService)
	authInjector := func(c *gin.Context) {
		c.Set("currentUser", adminUserForBackup)
		c.Set("userID", adminUserForBackup.ID)
		c.Next()
	}
	router := setupBackupTestRouter(mockBackupSvc, authInjector)

	// --- PERBAIKAN: Mengganti mock.AnythingOfType("io.Reader") menjadi mock.Anything ---
	// Ini akan menerima tipe spesifik 'multipart.sectionReadCloser'
	mockBackupSvc.On("RestoreBackup", mock.Anything, adminUserForBackup.ID).Return(nil).Once()
	// --- AKHIR PERBAIKAN ---

	// Buat body multipart form
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("restore-file", "backup.db")
	part.Write([]byte("dummy db data"))
	writer.Close()

	req, _ := http.NewRequest(http.MethodPost, "/api/restore", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.JSONEq(t, `{"message":"Database berhasil dipulihkan."}`, recorder.Body.String())
	mockBackupSvc.AssertExpectations(t)
}

// Test POST /api/restore - Failure (Wrong file extension)
func TestBackupController_RestoreBackup_WrongExtension(t *testing.T) {
	mockBackupSvc := new(mocks.BackupService)
	authInjector := func(c *gin.Context) {
		c.Set("currentUser", adminUserForBackup)
		c.Set("userID", adminUserForBackup.ID)
		c.Next()
	}
	router := setupBackupTestRouter(mockBackupSvc, authInjector)

	// (Mock service TIDAK akan dipanggil)

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("restore-file", "backup.txt") // Ekstensi .txt
	part.Write([]byte("dummy data"))
	writer.Close()

	req, _ := http.NewRequest(http.MethodPost, "/api/restore", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.JSONEq(t, `{"error":"Format file tidak valid. Harap unggah file .db"}`, recorder.Body.String())
	mockBackupSvc.AssertNotCalled(t, "RestoreBackup")
}