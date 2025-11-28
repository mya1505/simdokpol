package dto

type MigrationProgress struct {
	Step    string `json:"step"`    // Sedang memproses tabel apa
	Percent int    `json:"percent"` // 0-100
	Message string `json:"message"` // Pesan detail
}