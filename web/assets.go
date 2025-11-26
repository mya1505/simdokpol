package web

import (
	"embed"
	"net/http"
	"io/fs"
	"log"
)

//go:embed templates/*.html templates/partials/*.html static/*
var Assets embed.FS

// GetIconBytes membaca file icon.png dari embed FS dan mengembalikannya sebagai byte array
// Ini digunakan oleh Systray yang membutuhkan []byte
func GetIconBytes() []byte {
	// Sesuaikan path ini dengan lokasi fisik file lo: web/static/img/icon.png
	// Karena kita berada di package 'web', path relatifnya dari root embed
	iconData, err := Assets.ReadFile("static/img/icon.png")
	if err != nil {
		log.Printf("WARNING: Gagal memuat icon.png dari embed: %v", err)
		return nil
	}
	return iconData
}

// GetStaticFS mengembalikan sub-filesystem untuk folder static
// Ini agar Gin bisa serve "/static" tanpa prefix "static/" di URL
func GetStaticFS() http.FileSystem {
	sub, err := fs.Sub(Assets, "static")
	if err != nil {
		log.Fatal("Gagal membuat sub-fs untuk static:", err)
	}
	return http.FS(sub)
}
