/**
 * FILE HEADER: internal/controllers/helpers.go
 *
 * PURPOSE:
 * Menyediakan fungsi-fungsi helper untuk standarisasi respons API JSON.
 * Dengan menggunakan helper ini, semua respons (baik sukses maupun error)
 * akan memiliki format yang konsisten di seluruh aplikasi.
 */
package controllers

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

// APIResponse mengirimkan respons JSON standar untuk operasi yang sukses.
//
// PARAMETERS:
// - ctx (*gin.Context): Konteks request Gin.
// - statusCode (int): Kode status HTTP (misalnya, 200, 201).
// - message (string): Pesan singkat yang mendeskripsikan hasil.
// - data (interface{}): Payload data opsional yang akan disertakan dalam respons.
func APIResponse(ctx *gin.Context, statusCode int, message string, data interface{}) {
	response := gin.H{"message": message}
	if data != nil {
		response["data"] = data
	}
	ctx.JSON(statusCode, response)
}

// APIError mengirimkan respons JSON standar untuk error.
//
// PARAMETERS:
// - ctx (*gin.Context): Konteks request Gin.
// - statusCode (int): Kode status HTTP error (misalnya, 400, 403, 404, 500).
// - errorMessage (string): Pesan error yang aman untuk ditampilkan ke klien.
func APIError(ctx *gin.Context, statusCode int, errorMessage string) {
	ctx.JSON(statusCode, gin.H{"error": errorMessage})
}

// RenderHTML merender template HTML dengan menyertakan data global (seperti AppVersion dan CurrentUser)
func RenderHTML(ctx *gin.Context, templateName string, data gin.H) {
	// 1. Ambil AppVersion dari Context (yang diset di middleware main.go)
	if v, exists := ctx.Get("AppVersion"); exists {
		data["AppVersion"] = v
	} else {
		data["AppVersion"] = "dev" // Fallback
	}

	// 2. Ambil CurrentUser dari Context (jika ada, overwrite/inject)
	// Ini biar gak perlu manual kirim CurrentUser terus-terusan
	if u, exists := ctx.Get("currentUser"); exists {
		// Hanya set jika belum ada di data (biar fleksibel)
		if _, present := data["CurrentUser"]; !present {
			data["CurrentUser"] = u
		}
	}
	
    // 3. Ambil Changelog
    if cl, exists := ctx.Get("AppChangelog"); exists {
        data["AppChangelog"] = cl
    }

	ctx.HTML(http.StatusOK, templateName, data)
}