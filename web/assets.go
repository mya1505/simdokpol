package web

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"runtime"
	"sync"
)

//go:embed templates/*.html templates/partials/*.html static/*
var Assets embed.FS

var (
	cachedIconBytes []byte
	iconCacheMutex  sync.Mutex
	iconInitialized bool
)

// GetIconBytes membaca file icon dari embed FS dan mengembalikannya sebagai byte array
// Fungsi ini secara otomatis memilih format icon yang tepat berdasarkan sistem operasi:
// - Windows: menggunakan icon.ico (format ICO dengan multiple resolution)
// - Linux/macOS: menggunakan icon.png (format PNG)
// Icon di-cache setelah pembacaan pertama untuk optimasi performa
func GetIconBytes() []byte {
	iconCacheMutex.Lock()
	defer iconCacheMutex.Unlock()

	// Return cached icon jika sudah pernah dimuat
	if iconInitialized && cachedIconBytes != nil {
		return cachedIconBytes
	}

	var iconData []byte
	var err error
	var iconPath string

	// Pilih format icon berdasarkan sistem operasi
	if runtime.GOOS == "windows" {
		// Windows menggunakan ICO format untuk system tray yang optimal
		iconPath = "static/img/icon.ico"
		iconData, err = Assets.ReadFile(iconPath)
		
		if err != nil {
			// Fallback ke PNG jika ICO tidak tersedia
			log.Printf("⚠️ Icon ICO tidak ditemukan di %s: %v", iconPath, err)
			log.Println("⚠️ Menggunakan PNG sebagai fallback (mungkin tidak tampil optimal di Windows system tray)")
			iconPath = "static/img/icon.png"
			iconData, err = Assets.ReadFile(iconPath)
		} else {
			log.Printf("✓ Menggunakan icon ICO untuk Windows dari: %s", iconPath)
		}
	} else {
		// Unix-like systems (Linux, macOS) menggunakan PNG
		iconPath = "static/img/icon.png"
		iconData, err = Assets.ReadFile(iconPath)
		
		if err == nil {
			log.Printf("✓ Menggunakan icon PNG untuk %s dari: %s", runtime.GOOS, iconPath)
		}
	}

	// Handle error jika semua upaya gagal
	if err != nil {
		log.Printf("❌ GAGAL memuat icon dari embed FS: %v", err)
		log.Println("⚠️ System tray icon mungkin tidak akan ditampilkan")
		return nil
	}

	// Cache icon untuk penggunaan selanjutnya
	cachedIconBytes = iconData
	iconInitialized = true

	return iconData
}

// GetIconFormat mengembalikan format icon yang sedang digunakan
// Berguna untuk debugging dan logging
func GetIconFormat() string {
	if runtime.GOOS == "windows" {
		// Cek apakah ICO tersedia
		_, err := Assets.ReadFile("static/img/icon.ico")
		if err == nil {
			return "ICO (Windows native format)"
		}
		return "PNG (fallback - ICO not available)"
	}
	return "PNG (native format for " + runtime.GOOS + ")"
}

// GetIconInfo mengembalikan informasi lengkap tentang konfigurasi icon
// Berguna untuk troubleshooting dan monitoring
func GetIconInfo() map[string]interface{} {
	info := make(map[string]interface{})
	
	info["os"] = runtime.GOOS
	info["format"] = GetIconFormat()
	info["cache_initialized"] = iconInitialized
	
	// Cek ketersediaan file icon
	icoAvailable := false
	pngAvailable := false
	
	if _, err := Assets.ReadFile("static/img/icon.ico"); err == nil {
		icoAvailable = true
	}
	
	if _, err := Assets.ReadFile("static/img/icon.png"); err == nil {
		pngAvailable = true
	}
	
	info["ico_available"] = icoAvailable
	info["png_available"] = pngAvailable
	
	if cachedIconBytes != nil {
		info["cached_size_bytes"] = len(cachedIconBytes)
	} else {
		info["cached_size_bytes"] = 0
	}
	
	return info
}

// GetStaticFS mengembalikan sub-filesystem untuk folder static
// Ini agar Gin bisa serve "/static" tanpa prefix "static/" di URL
func GetStaticFS() http.FileSystem {
	sub, err := fs.Sub(Assets, "static")
	if err != nil {
		log.Fatal("❌ Gagal membuat sub-fs untuk static:", err)
	}
	return http.FS(sub)
}