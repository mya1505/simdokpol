package utils

import (
	"bytes"
	"fmt"
	"log"
	"simdokpol/internal/dto"
	"simdokpol/internal/models"
	"simdokpol/web" // Import package web (Embed)
	"strings"

	"github.com/jung-kurt/gofpdf"
)

// GenerateLostDocumentPDF membuat PDF berdasarkan data dokumen dan konfigurasi
func GenerateLostDocumentPDF(doc *models.LostDocument, config *dto.AppConfig, exeDir string) (*bytes.Buffer, string) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetMargins(20, 15, 20) // Atur margin Kiri, Atas, Kanan
	pdf.SetAutoPageBreak(true, 15) // Margin Bawah

	const (
		lineHeight       = 4.2
		lineHeightKop    = 3.8
		spasiTtd         = 10
		spasiTtdBermohon = 10
		lineHeightItem   = 4
	)

	// --- LOAD LOGO DARI EMBED (SINGLE BINARY SAFE) ---
	logoBytes, err := web.Assets.ReadFile("static/img/logo.png")
	if err == nil {
		// Register gambar dari memory ke PDF engine
		logoReader := bytes.NewReader(logoBytes)
		pdf.RegisterImageOptionsReader("logo.png", gofpdf.ImageOptions{ImageType: "PNG"}, logoReader)
	} else {
		log.Printf("WARNING: Gagal load logo dari embed: %v", err)
	}

	// Helper untuk styling font
	setFont := func(style string, size float64) {
		pdf.SetFont("Courier", style, size)
	}
	
	// Helper untuk text-align center halaman penuh
	cellFitCenter := func(h float64, text string) {
		pageWidth, _ := pdf.GetPageSize()
		marginL, _, marginR, _ := pdf.GetMargins()
		printableWidth := pageWidth - marginL - marginR
		strWidth := pdf.GetStringWidth(text)
		
		// Hitung X agar tengah
		x := marginL + (printableWidth - strWidth) / 2
		pdf.SetX(x) 
		pdf.Cell(strWidth, h, text)
	}

	// --- 1. KOP SURAT ---
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

	// --- 2. LOGO ---
	if err == nil {
		pdf.Image("logo.png", 99, 42, 12, 0, false, "", 0, "")
	}
	pdf.SetY(56)
	
	// --- 3. JUDUL DOKUMEN ---
	setFont("BU", 12)
	cellFitCenter(lineHeight, "SURAT KETERANGAN HILANG")
	pdf.Ln(lineHeight)
	setFont("", 11)
	cellFitCenter(lineHeight, fmt.Sprintf("Nomor: %s", doc.NomorSurat))
	pdf.Ln(8)

	// --- 4. PARAGRAF PEMBUKA ---
	setFont("", 11)
	pdf.SetX(20)
	pdf.MultiCell(170, lineHeight, fmt.Sprintf("---- Yang bertanda tangan dibawah ini A.n. KEPALA KEPOLISIAN %s, Menerangkan dengan benar bahwa :", strings.ToUpper(config.KopBaris3)), "", "J", false)
	pdf.Ln(lineHeight)

	// --- 5. DATA PEMOHON ---
	pdf.SetX(30)
	pdf.Cell(40, lineHeight, "Nama")
	pdf.Cell(5, lineHeight, ":")
	setFont("B", 11)
	pdf.Cell(0, lineHeight, strings.ToUpper(doc.Resident.NamaLengkap))
	pdf.Ln(lineHeight)
	
	setFont("", 11)
	pdf.SetX(30)
	pdf.Cell(40, lineHeight, "TTL")
	pdf.Cell(5, lineHeight, ":")
	pdf.Cell(0, lineHeight, fmt.Sprintf("%s, %s", doc.Resident.TempatLahir, doc.Resident.TanggalLahir.Format("02-01-2006")))
	pdf.Ln(lineHeight)

	pdf.SetX(30)
	pdf.Cell(40, lineHeight, "Agama")
	pdf.Cell(5, lineHeight, ":")
	pdf.Cell(0, lineHeight, doc.Resident.Agama)
	pdf.Ln(lineHeight)

	pdf.SetX(30)
	pdf.Cell(40, lineHeight, "Jenis kelamin")
	pdf.Cell(5, lineHeight, ":")
	pdf.Cell(0, lineHeight, doc.Resident.JenisKelamin)
	pdf.Ln(lineHeight)

	pdf.SetX(30)
	pdf.Cell(40, lineHeight, "Pekerjaan")
	pdf.Cell(5, lineHeight, ":")
	pdf.Cell(0, lineHeight, doc.Resident.Pekerjaan)
	pdf.Ln(lineHeight)

	pdf.SetX(30)
	pdf.Cell(40, lineHeight, "Alamat")
	pdf.Cell(5, lineHeight, ":")
	pdf.MultiCell(115, lineHeight, doc.Resident.Alamat, "", "L", false)
	pdf.Ln(3)

	// --- 6. PARAGRAF TENGAH (BARANG HILANG) ---
	pdf.SetX(20)
	pdf.MultiCell(170, lineHeight, fmt.Sprintf("Yang bersangkutan tersebut di atas benar telah datang di Kantor %s dan melaporkan bahwa telah kehilangan surat berharga berupa :", config.NamaKantor), "", "J", false)
	pdf.Ln(3)

	// Daftar barang
	for i, item := range doc.LostItems {
		pdf.SetX(30)
		line := fmt.Sprintf("- 1 (Buah) %s Dengan Keterangan : %s A.n Pelapor", strings.ToUpper(item.NamaBarang), item.Deskripsi)
		pdf.MultiCell(160, lineHeightItem, line, "", "L", false)
		
		if i < len(doc.LostItems)-1 {
			pdf.Ln(0.5)
		}
	}
	pdf.Ln(3)

	// --- 7. PARAGRAF PENUTUP (Lokasi Hilang) ---
	pdf.SetX(20)
	pdf.MultiCell(170, lineHeight, fmt.Sprintf("---- Surat/kartu tersebut hilang di sekitar %s, dan sudah dilakukan pencarian namun sampai dikeluarkan Surat Keterangan ini belum ditemukan.", doc.LokasiHilang), "", "J", false)
	pdf.Ln(3)

	// --- 8. BLOK YANG BERMOHON (Kanan) ---
	pdf.SetX(120)
	pdf.MultiCell(70, lineHeight, "Yang Bermohon", "", "C", false)
	pdf.Ln(spasiTtdBermohon)
	pdf.SetX(120)
	setFont("BU", 11)
	pdf.MultiCell(70, lineHeight, strings.ToUpper(doc.Resident.NamaLengkap), "", "C", false)
	yPosAfterBermohon := pdf.GetY()

	// --- 9. PARAGRAF PENUTUP 2 ---
	setFont("", 11)
	pdf.SetY(yPosAfterBermohon)
	pdf.Ln(3)
	pdf.SetX(20)
	pdf.MultiCell(170, lineHeight, "----- Demikian Surat Keterangan ini dibuat dengan sebenar-benarnya dan dapat dipergunakan sebagaimana perlunya.", "", "J", false)
	yPosAfterDemikian := pdf.GetY()

	// --- 10. TINDAKAN YANG DIAMBIL ---
	pdf.SetY(yPosAfterDemikian)
	pdf.Ln(3)

	setFont("BU", 11)
	pdf.SetX(20)
	pdf.Cell(0, lineHeight, "Tindakan Yang Diambil :")
	pdf.Ln(lineHeight)

	setFont("", 11)
	pdf.SetX(25)
	pdf.MultiCell(165, lineHeightItem, "1. Menerima laporan dan membuat Surat Keterangan Kehilangan barang guna seperlunya;", "", "L", false)
	pdf.SetX(25)

	archiveDays := 15
	if config.ArchiveDurationDays > 0 {
		archiveDays = config.ArchiveDurationDays
	}
	
	archiveDaysWords := IntToIndonesianWords(archiveDays)
	pdf.MultiCell(165, lineHeightItem, fmt.Sprintf("2. Surat keterangan kehilangan ini berlaku selama %d (%s) hari, berlaku mulai tanggal dikeluarkan;", archiveDays, archiveDaysWords), "", "L", false)

	pdf.SetX(25)
	pdf.MultiCell(165, lineHeightItem, "3. Surat Keterangan ini bukan sebagai pengganti surat yang hilang tetapi berguna untuk mengurus kembali surat yang hilang.", "", "L", false)
	pdf.Ln(3)

	// --- 11. BLOK TANDA TANGAN BAWAH ---
	
	// Tanggal
	pdf.SetX(120)
	pdf.MultiCell(70, lineHeight, fmt.Sprintf("%s, %s", config.TempatSurat, doc.TanggalLaporan.Format("02 January 2006")), "", "C", false)
	pdf.Ln(2)

	yPosTtd := pdf.GetY()
	
	jabatanPersetuju := strings.ToUpper(doc.PejabatPersetuju.Jabatan)
	if doc.PejabatPersetuju.Regu != "" && doc.PejabatPersetuju.Regu != "-" {
		jabatanPersetuju = fmt.Sprintf("%s %s", jabatanPersetuju, doc.PejabatPersetuju.Regu)
	}

	jabatanPelapor := strings.ToUpper(doc.PetugasPelapor.Jabatan)
	if doc.PetugasPelapor.Regu != "" && doc.PetugasPelapor.Regu != "-" {
		jabatanPelapor = fmt.Sprintf("%s %s", jabatanPelapor, doc.PetugasPelapor.Regu)
	}

	// Kolom Kiri (Pejabat Persetuju)
	pdf.SetY(yPosTtd)
	pdf.SetX(20)
	setFont("B", 9)
	pdf.MultiCell(85, lineHeightKop, fmt.Sprintf("a.n. KEPALA KEPOLISIAN %s", strings.ToUpper(config.KopBaris3)), "", "C", false)
	pdf.SetX(20)
	pdf.MultiCell(85, lineHeightKop, jabatanPersetuju, "", "C", false)
	pdf.Ln(spasiTtd)
	pdf.SetX(20)
	setFont("BU", 9)
	pdf.MultiCell(85, lineHeightKop, strings.ToUpper(doc.PejabatPersetuju.NamaLengkap), "", "C", false)
	setFont("", 9)
	pdf.SetX(20)
	pdf.MultiCell(85, lineHeightKop, fmt.Sprintf("%s / NRP %s", strings.ToUpper(doc.PejabatPersetuju.Pangkat), doc.PejabatPersetuju.NRP), "", "C", false)

	// Kolom Kanan (Penerima Laporan)
	pdf.SetY(yPosTtd)
	pdf.SetX(120)
	setFont("B", 9)
	pdf.MultiCell(70, lineHeightKop, "Penerima Laporan", "", "C", false)
	pdf.SetX(120)
	pdf.MultiCell(70, lineHeightKop, jabatanPelapor, "", "C", false)
	pdf.Ln(spasiTtd)
	pdf.SetX(120)
	setFont("BU", 9)
	pdf.MultiCell(70, lineHeightKop, strings.ToUpper(doc.PetugasPelapor.NamaLengkap), "", "C", false)
	setFont("", 9)
	pdf.SetX(120)
	pdf.MultiCell(70, lineHeightKop, fmt.Sprintf("%s / NRP %s", strings.ToUpper(doc.PetugasPelapor.Pangkat), doc.PetugasPelapor.NRP), "", "C", false)

	// --- FINALISASI ---
	var buffer bytes.Buffer
	err = pdf.Output(&buffer)
	if err != nil {
		return nil, ""
	}

	filename := fmt.Sprintf("SKH_%s_%s.pdf", strings.ReplaceAll(doc.Resident.NamaLengkap, " ", "_"), doc.TanggalLaporan.Format("20060102"))
	return &buffer, filename
}