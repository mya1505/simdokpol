package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"simdokpol/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSystemController_HealthzDegradedWithoutDB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	controller := NewSystemController(nil)

	router := gin.New()
	router.GET("/api/healthz", controller.Healthz)

	req, _ := http.NewRequest(http.MethodGet, "/api/healthz", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"degraded"`)
}

func TestSystemController_MetricsWithDB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("gagal membuka sqlite memory: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Resident{}, &models.LostDocument{}, &models.LostItem{}, &models.AuditLog{}, &models.ItemTemplate{}); err != nil {
		t.Fatalf("gagal migrate: %v", err)
	}

	controller := NewSystemController(db)
	router := gin.New()
	router.GET("/api/metrics", controller.Metrics)

	req, _ := http.NewRequest(http.MethodGet, "/api/metrics", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var payload map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("gagal decode response: %v", err)
	}
	assert.Contains(t, payload, "documents")
	assert.Contains(t, payload, "users")
	assert.Contains(t, payload, "uptime_s")
}
