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

func TestBackupController_CreateBackup(t *testing.T) {
	mockBackupSvc := new(mocks.BackupService)
	authInjector := func(c *gin.Context) {
		c.Set("currentUser", adminUserForBackup)
		c.Set("userID", adminUserForBackup.ID)
		c.Next()
	}
	router := setupBackupTestRouter(mockBackupSvc, authInjector)

	tempDir := t.TempDir()
	dummyFileName := "test_backup.db"
	dummyFilePath := filepath.Join(tempDir, dummyFileName)
	err := os.WriteFile(dummyFilePath, []byte("dummy db data"), 0644)
	assert.NoError(t, err)
	
	mockBackupSvc.On("CreateBackup", adminUserForBackup.ID).Return(dummyFilePath, nil).Once()

	req, _ := http.NewRequest(http.MethodPost, "/api/backups", nil)
	req.Header.Set("Accept", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	
	// FIX: Controller saat ini tidak pakai quotes di filename, sesuaikan testnya
	expectedHeader := `attachment; filename=` + dummyFileName
	assert.Contains(t, recorder.Header().Get("Content-Disposition"), expectedHeader)

	assert.Equal(t, "dummy db data", recorder.Body.String())
	mockBackupSvc.AssertExpectations(t)
}

func TestBackupController_RestoreBackup(t *testing.T) {
	mockBackupSvc := new(mocks.BackupService)
	authInjector := func(c *gin.Context) {
		c.Set("currentUser", adminUserForBackup)
		c.Set("userID", adminUserForBackup.ID)
		c.Next()
	}
	router := setupBackupTestRouter(mockBackupSvc, authInjector)

	mockBackupSvc.On("RestoreBackup", mock.Anything, adminUserForBackup.ID).Return(nil).Once()

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
	
	// FIX: Update pesan sukses sesuai implementasi controller terbaru (Auto Restart)
	assert.JSONEq(t, `{"message":"Restore berhasil. Silakan restart aplikasi."}`, recorder.Body.String())
	mockBackupSvc.AssertExpectations(t)
}

func TestBackupController_RestoreBackup_WrongExtension(t *testing.T) {
	mockBackupSvc := new(mocks.BackupService)
	authInjector := func(c *gin.Context) {
		c.Set("currentUser", adminUserForBackup)
		c.Set("userID", adminUserForBackup.ID)
		c.Next()
	}
	router := setupBackupTestRouter(mockBackupSvc, authInjector)

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("restore-file", "backup.txt")
	part.Write([]byte("dummy data"))
	writer.Close()

	req, _ := http.NewRequest(http.MethodPost, "/api/restore", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	
	// FIX: Update pesan error sesuai implementasi controller terbaru
	assert.JSONEq(t, `{"error":"Format harus .db"}`, recorder.Body.String())
	mockBackupSvc.AssertNotCalled(t, "RestoreBackup")
}