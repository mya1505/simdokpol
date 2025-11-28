package middleware

import (
	"net/http"
	"simdokpol/internal/services"

	"github.com/gin-gonic/gin"
)

func LicenseMiddleware(licenseService services.LicenseService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if licenseService.IsLicensed() {
			c.Next()
			return
		}

		// Jika Lisensi Tidak Valid / Expired
		
		// Cek apakah request minta JSON (API Call)
		if c.Request.Header.Get("Accept") == "application/json" {
			c.JSON(http.StatusPaymentRequired, gin.H{
                "error": "Fitur ini terkunci. Silakan upgrade ke versi Pro.",
                "upgrade_url": "/upgrade",
            })
		} else {
			// Jika request Halaman Web biasa, Redirect ke halaman promosi
			c.Redirect(http.StatusFound, "/upgrade")
		}
		c.Abort()
	}
}