package dto

import (
	"simdokpol/internal/models"
	"time"
)

// --- Definisi dari fitur Laporan Agregat (tetap ada) ---
type ItemCompositionStat struct {
	NamaBarang string `gorm:"column:nama_barang"`
	Count      int    `gorm:"column:count"`
}

type OperatorStat struct {
	OperatorID  uint   `gorm:"column:operator_id"`
	NamaLengkap string `gorm:"column:nama_lengkap"`
	Count       int    `gorm:"column:count"`
}

type AggregateReportData struct {
	StartDate       time.Time
	EndDate         time.Time
	TotalDocuments  int64
	ItemComposition []ItemCompositionStat
	OperatorStats   []OperatorStat
	DocumentList    []models.LostDocument
}

// --- STRUCT BARU UNTUK TES KONEKSI DB ---
type DBTestRequest struct {
	DBDialect string `json:"db_dialect" binding:"required"`
	DBHost    string `json:"db_host"`
	DBPort    string `json:"db_port"`
	DBUser    string `json:"db_user"`
	DBPass    string `json:"db_pass"`
	DBName    string `json:"db_name"`
	DBSSLMode string `json:"db_sslmode"` // <-- TAMBAHAN BARU
}