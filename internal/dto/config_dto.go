package dto

// AppConfig adalah Data Transfer Object untuk konfigurasi aplikasi.
type AppConfig struct {
	IsSetupComplete     bool   `json:"is_setup_complete"`
	KopBaris1           string `json:"kop_baris_1"`
	KopBaris2           string `json:"kop_baris_2"`
	KopBaris3           string `json:"kop_baris_3"`
	NamaKantor          string `json:"nama_kantor"`
	TempatSurat         string `json:"tempat_surat"`
	FormatNomorSurat    string `json:"format_nomor_surat"`
	NomorSuratTerakhir  string `json:"nomor_surat_terakhir"`
	ZonaWaktu           string `json:"zona_waktu"`
	BackupPath          string `json:"backup_path"`
	ArchiveDurationDays int    `json:"archive_duration_days"`
	
	// --- PENTING: Field ini wajib ada biar settingan HTTPS kebaca ---
	EnableHTTPS         bool   `json:"enable_https"` 
	// --------------------------------------------------------------

	DBDialect           string `json:"db_dialect"`
	DBHost              string `json:"db_host"`
	DBPort              string `json:"db_port"`
	DBUser              string `json:"db_user"`
	DBName              string `json:"db_name"`
	DBDSN               string `json:"db_dsn"`
	DBSSLMode           string `json:"db_sslmode"`
	LicenseStatus       string `json:"license_status"`
}