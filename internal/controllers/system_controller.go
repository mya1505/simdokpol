package controllers

import (
	"net/http"
	"simdokpol/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SystemController struct {
	db        *gorm.DB
	startedAt time.Time
}

func NewSystemController(db *gorm.DB) *SystemController {
	return &SystemController{
		db:        db,
		startedAt: time.Now(),
	}
}

func (c *SystemController) Healthz(ctx *gin.Context) {
	status := "ok"
	dbStatus := "ready"

	if c.db == nil {
		status = "degraded"
		dbStatus = "unavailable"
	} else if sqlDB, err := c.db.DB(); err != nil || sqlDB.Ping() != nil {
		status = "degraded"
		dbStatus = "error"
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":    status,
		"db_status": dbStatus,
		"uptime_s":  int64(time.Since(c.startedAt).Seconds()),
	})
}

func (c *SystemController) Metrics(ctx *gin.Context) {
	if c.db == nil {
		APIError(ctx, http.StatusServiceUnavailable, "Database tidak tersedia.")
		return
	}

	var users, residents, documents, items, logs, templates int64
	c.db.Model(&models.User{}).Count(&users)
	c.db.Model(&models.Resident{}).Count(&residents)
	c.db.Model(&models.LostDocument{}).Count(&documents)
	c.db.Model(&models.LostItem{}).Count(&items)
	c.db.Model(&models.AuditLog{}).Count(&logs)
	c.db.Model(&models.ItemTemplate{}).Count(&templates)

	ctx.JSON(http.StatusOK, gin.H{
		"uptime_s":   int64(time.Since(c.startedAt).Seconds()),
		"users":      users,
		"residents":  residents,
		"documents":  documents,
		"lost_items": items,
		"audit_logs": logs,
		"templates":  templates,
	})
}
