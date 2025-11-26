package middleware

import (
	"net/http"
	"simdokpol/internal/services"

	"github.com/gin-gonic/gin"
)

// LicenseMiddleware memeriksa apakah lisensi aplikasi valid.
// Jika tidak, blokir akses ke fitur-fitur "Pro".
func LicenseMiddleware(licenseService services.LicenseService) gin.HandlerFunc {
	return func(c *gin.Context) {
		
		if licenseService.IsLicensed() {
			// Lisensi valid, lanjutkan ke controller
			c.Next()
			return
		}

		// Lisensi tidak valid, blokir request
		
		// Jika request adalah API (minta JSON)
		if c.Request.Header.Get("Accept") == "application/json" {
			c.JSON(http.StatusPaymentRequired, gin.H{"error": "Fitur Pro. Harap aktifkan lisensi Anda untuk menggunakan fitur ini."})
		} else {
			// Jika request adalah halaman web (minta HTML)
			c.HTML(http.StatusPaymentRequired, "error.html", gin.H{
				"Title": "Fitur Pro",
				"ErrorMessage": "Fitur ini memerlukan lisensi Pro yang aktif. Silakan aktifkan lisensi Anda di halaman Pengaturan.",
			})
		}
		c.Abort()
	}
}