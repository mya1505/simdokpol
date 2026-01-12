package utils

import "time"

// FormatTanggalIndonesia returns a date formatted with Indonesian month names.
func FormatTanggalIndonesia(t time.Time) string {
	bulan := []string{
		"Januari", "Februari", "Maret", "April", "Mei", "Juni",
		"Juli", "Agustus", "September", "Oktober", "November", "Desember",
	}
	monthIndex := int(t.Month()) - 1
	if monthIndex < 0 || monthIndex >= len(bulan) {
		return t.Format("02 January 2006")
	}
	return t.Format("02 ") + bulan[monthIndex] + t.Format(" 2006")
}
