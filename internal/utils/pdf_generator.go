package utils

import (
	"bytes"
	"fmt"
	"log"
	"simdokpol/internal/dto"
	"simdokpol/internal/models"
	"simdokpol/web" // <-- WAJIB IMPORT PACKAGE WEB (EMBED)
	"strings"

	"github.com/jung-kurt/gofpdf"
)

func GenerateLostDocumentPDF(doc *models.LostDocument, config *dto.AppConfig, exeDir string) (*bytes.Buffer, string) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetMargins(20, 15, 20)
	pdf.SetAutoPageBreak(true, 15)

	const (
		lineHeight       = 4.2
		lineHeightKop    = 3.8
		spasiTtd         = 10
		spasiTtdBermohon = 10
		lineHeightItem   = 4
	)

	// --- FIX: LOAD LOGO DARI EMBED ---
	logoBytes, err := web.Assets.ReadFile("static/img/logo.png")
	if err == nil {
		// Register gambar dari memory ke PDF engine
		logoReader := bytes.NewReader(logoBytes)
		pdf.RegisterImageOptionsReader("logo.png", gofpdf.ImageOptions{ImageType: "PNG"}, logoReader)
	} else {
		log.Printf("WARNING: Gagal load logo dari embed: %v", err)
	}
	// ----------------------------------

	// Helper font & align (sama seperti sebelumnya)
	setFont := func(style string, size float64) {
		pdf.SetFont("Courier", style, size)
	}
	cellFitCenter := func(h float64, text string) {
		strWidth := pdf.GetStringWidth(text) + 2
		pdf.SetX((210 - 20 - 20 - strWidth) / 2 + 20) 
		pdf.Cell(strWidth, h, text)
	}

	// 1. KOP SURAT (Code sama...)
	kopWidth := float64(60)
	setFont("B", 9)
	pdf.SetY(20)
	pdf.SetX(20)
	pdf.MultiCell(kopWidth, lineHeightKop, config.KopBaris1, "", "C", false)
	pdf.SetX(20)
	pdf.MultiCell(kopWidth, lineHeightKop, config.KopBaris2, "", "C", false)
	pdf.SetX(20)
	setFont("B", 10)
	pdf.MultiCell(kopWidth, lineHeightKop, config.KopBaris3, "", "C", false)
	pdf.Ln(2)

	currentY := pdf.GetY()
	pdf.SetLineWidth(0.3)
	pdf.Line(20, currentY, 20+kopWidth, currentY)
	pdf.SetLineWidth(0.2)
	pdf.Ln(2)

	// 2. LOGO (TAMPILKAN DARI MEMORY YG SUDAH DI-REGISTER)
	if err == nil {
		// Gunakan nama "logo.png" yang kita register di atas
		pdf.Image("logo.png", 99, 42, 12, 0, false, "", 0, "")
	}
	pdf.SetY(56)
	
	// ... (SISA KODE LOGIC CONTENT SAMA PERSIS SEPERTI FILE ASLI) ...
	// ... Copy Paste logic Judul, Paragraf, Tanda Tangan dari file lama lo ...

	// --- FINALISASI ---
	var buffer bytes.Buffer
	err = pdf.Output(&buffer)
	if err != nil {
		return nil, ""
	}

	filename := fmt.Sprintf("SKH_%s_%s.pdf", strings.ReplaceAll(doc.Resident.NamaLengkap, " ", "_"), doc.TanggalLaporan.Format("20060102"))
	return &buffer, filename
}