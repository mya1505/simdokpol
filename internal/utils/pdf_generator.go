package utils

import (
	"bytes"
	"fmt"
	"log"
	"math"
	"simdokpol/internal/dto"
	"simdokpol/internal/models"
	"simdokpol/web"
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
		spasiTtd         = 15
		spasiTtdBermohon = 15
		lineHeightItem   = 4
		logoX            = 99.0
	)

	// --- LOAD LOGO ---
	logoBytes, err := web.Assets.ReadFile("static/img/logo.png")
	if err == nil {
		logoReader := bytes.NewReader(logoBytes)
		pdf.RegisterImageOptionsReader("logo.png", gofpdf.ImageOptions{ImageType: "PNG"}, logoReader)
	} else {
		log.Printf("WARNING: Gagal load logo: %v", err)
	}

	setFont := func(style string, size float64) { pdf.SetFont("Courier", style, size) }
	
	// Helper untuk Hitung Lebar Text Terpanjang di Blok TTD
	calcMaxW := func(lines ...string) float64 {
		setFont("B", 9) // Asumsi font terbesar di TTD
		max := 0.0
		for _, l := range lines {
			if w := pdf.GetStringWidth(l); w > max { max = w }
		}
		return max + 2.0 // Padding
	}

	// --- 1. KOP SURAT (DINAMIS) ---
	setFont("B", 10)
	w1 := pdf.GetStringWidth(config.KopBaris1)
	w2 := pdf.GetStringWidth(config.KopBaris2)
	w3 := pdf.GetStringWidth(config.KopBaris3)
	maxTextWidth := math.Max(w1, math.Max(w2, w3))
	kopWidth := maxTextWidth + 4.0
	if kopWidth < 60.0 { kopWidth = 60.0 }
	if kopWidth > 75.0 { kopWidth = 75.0 }

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
	if err == nil { pdf.Image("logo.png", logoX, 42, 12, 0, false, "", 0, "") }
	
	pdf.SetY(56)
	
	// --- 3. BODY (JUDUL & ISI) ---
	// ... (Bagian Body ini sama seperti sebelumnya, dipersingkat) ...
	// Gunakan kode sebelumnya untuk bagian Judul s.d. Tindakan Yang Diambil
	
	setFont("BU", 12)
	pageWidth, _ := pdf.GetPageSize()
	marginL, _, marginR, _ := pdf.GetMargins()
	printableWidth := pageWidth - marginL - marginR
	
	// Helper Center
	center := func(txt string) {
		w := pdf.GetStringWidth(txt)
		pdf.SetX(marginL + (printableWidth - w)/2)
		pdf.Cell(w, lineHeight, txt)
	}

	center("SURAT KETERANGAN HILANG")
	pdf.Ln(lineHeight)
	setFont("", 11)
	center(fmt.Sprintf("Nomor: %s", doc.NomorSurat))
	pdf.Ln(8)

	setFont("", 11)
	pdf.SetX(20)
	pdf.MultiCell(170, lineHeight, fmt.Sprintf("---- Yang bertanda tangan dibawah ini A.n. KEPALA KEPOLISIAN %s, Menerangkan dengan benar bahwa :", strings.ToUpper(config.KopBaris3)), "", "J", false)
	pdf.Ln(lineHeight)

	// Data Pemohon
	pdf.SetX(30); pdf.Cell(40, lineHeight, "Nama"); pdf.Cell(5, lineHeight, ":"); setFont("B", 11); pdf.Cell(0, lineHeight, strings.ToUpper(doc.Resident.NamaLengkap)); pdf.Ln(lineHeight)
	setFont("", 11)
	pdf.SetX(30); pdf.Cell(40, lineHeight, "TTL"); pdf.Cell(5, lineHeight, ":"); pdf.Cell(0, lineHeight, fmt.Sprintf("%s, %s", doc.Resident.TempatLahir, doc.Resident.TanggalLahir.Format("02-01-2006"))); pdf.Ln(lineHeight)
	pdf.SetX(30); pdf.Cell(40, lineHeight, "Agama"); pdf.Cell(5, lineHeight, ":"); pdf.Cell(0, lineHeight, doc.Resident.Agama); pdf.Ln(lineHeight)
	pdf.SetX(30); pdf.Cell(40, lineHeight, "Jenis kelamin"); pdf.Cell(5, lineHeight, ":"); pdf.Cell(0, lineHeight, doc.Resident.JenisKelamin); pdf.Ln(lineHeight)
	pdf.SetX(30); pdf.Cell(40, lineHeight, "Pekerjaan"); pdf.Cell(5, lineHeight, ":"); pdf.Cell(0, lineHeight, doc.Resident.Pekerjaan); pdf.Ln(lineHeight)
	pdf.SetX(30); pdf.Cell(40, lineHeight, "Alamat"); pdf.Cell(5, lineHeight, ":"); pdf.MultiCell(115, lineHeight, doc.Resident.Alamat, "", "L", false); pdf.Ln(3)

	// Barang Hilang
	pdf.SetX(20)
	pdf.MultiCell(170, lineHeight, fmt.Sprintf("Yang bersangkutan tersebut di atas benar telah datang di Kantor %s dan melaporkan bahwa telah kehilangan surat berharga berupa :", config.NamaKantor), "", "J", false)
	pdf.Ln(3)
	for i, item := range doc.LostItems {
		pdf.SetX(30)
		pdf.MultiCell(160, lineHeightItem, fmt.Sprintf("- 1 (Buah) %s Dengan Keterangan : %s A.n Pelapor", strings.ToUpper(item.NamaBarang), item.Deskripsi), "", "L", false)
		if i < len(doc.LostItems)-1 { pdf.Ln(0.5) }
	}
	pdf.Ln(3)

	// Penutup
	pdf.SetX(20)
	pdf.MultiCell(170, lineHeight, fmt.Sprintf("---- Surat/kartu tersebut hilang di sekitar %s, dan sudah dilakukan pencarian namun sampai dikeluarkan Surat Keterangan ini belum ditemukan.", doc.LokasiHilang), "", "J", false)
	pdf.Ln(3)

	// TTD Pemohon
	pdf.SetX(120)
	pdf.MultiCell(70, lineHeight, "Yang Bermohon", "", "C", false)
	pdf.Ln(spasiTtdBermohon)
	pdf.SetX(120)
	setFont("BU", 11)
	pdf.MultiCell(70, lineHeight, strings.ToUpper(doc.Resident.NamaLengkap), "", "C", false)
	
	// Demikian & Tindakan
	setFont("", 11)
	pdf.Ln(3)
	pdf.SetX(20)
	pdf.MultiCell(170, lineHeight, "----- Demikian Surat Keterangan ini dibuat dengan sebenar-benarnya dan dapat dipergunakan sebagaimana perlunya.", "", "J", false)
	pdf.Ln(3)
	setFont("BU", 11)
	pdf.SetX(20); pdf.Cell(0, lineHeight, "Tindakan Yang Diambil :"); pdf.Ln(lineHeight)
	setFont("", 11)
	pdf.SetX(25); pdf.MultiCell(165, lineHeightItem, "1. Menerima laporan dan membuat Surat Keterangan Kehilangan barang guna seperlunya;", "", "L", false)
	
	archiveDays := 15
	if config.ArchiveDurationDays > 0 { archiveDays = config.ArchiveDurationDays }
	pdf.SetX(25); pdf.MultiCell(165, lineHeightItem, fmt.Sprintf("2. Surat keterangan kehilangan ini berlaku selama %d (%s) hari, berlaku mulai tanggal dikeluarkan;", archiveDays, IntToIndonesianWords(archiveDays)), "", "L", false)
	pdf.SetX(25); pdf.MultiCell(165, lineHeightItem, "3. Surat Keterangan ini bukan sebagai pengganti surat yang hilang tetapi berguna untuk mengurus kembali surat yang hilang.", "", "L", false)
	pdf.Ln(3)

	// --- 4. TANDA TANGAN (DINAMIS WIDTH) ---
	
	// Siapkan Data Teks
	dateStr := fmt.Sprintf("%s, %s", config.TempatSurat, doc.TanggalLaporan.Format("02 January 2006"))
	
	// KIRI: PEJABAT
	jabatanKiri := fmt.Sprintf("a.n. KEPALA KEPOLISIAN %s", strings.ToUpper(config.KopBaris3))
	jabatanKiri2 := strings.ToUpper(doc.PejabatPersetuju.Jabatan)
	if doc.PejabatPersetuju.Regu != "" && doc.PejabatPersetuju.Regu != "-" {
		jabatanKiri2 += " " + doc.PejabatPersetuju.Regu
	}
	namaKiri := strings.ToUpper(doc.PejabatPersetuju.NamaLengkap)
	nrpKiri := fmt.Sprintf("%s / NRP %s", strings.ToUpper(doc.PejabatPersetuju.Pangkat), doc.PejabatPersetuju.NRP)

	// KANAN: PELAPOR
	jabatanKanan := "Penerima Laporan"
	jabatanKanan2 := strings.ToUpper(doc.PetugasPelapor.Jabatan)
	if doc.PetugasPelapor.Regu != "" && doc.PetugasPelapor.Regu != "-" {
		jabatanKanan2 += " " + doc.PetugasPelapor.Regu
	}
	namaKanan := strings.ToUpper(doc.PetugasPelapor.NamaLengkap)
	nrpKanan := fmt.Sprintf("%s / NRP %s", strings.ToUpper(doc.PetugasPelapor.Pangkat), doc.PetugasPelapor.NRP)

	// Hitung Lebar Kolom Dinamis
	// Min 70mm, Max 90mm biar gak tabrakan
	widthKiri := calcMaxW(jabatanKiri, jabatanKiri2, namaKiri, nrpKiri)
	if widthKiri < 70 { widthKiri = 70 }
	if widthKiri > 90 { widthKiri = 90 }

	widthKanan := calcMaxW(dateStr, jabatanKanan, jabatanKanan2, namaKanan, nrpKanan)
	if widthKanan < 70 { widthKanan = 70 }
	if widthKanan > 90 { widthKanan = 90 }

	// Tentukan Posisi X (Kiri Fix di 20, Kanan disesuaikan agar margin kanan 20)
	xKiri := 20.0
	xKanan := pageWidth - marginR - widthKanan // Align Right terhadap Margin

	// Render TTD
	// Tanggal (Kanan Saja)
	pdf.SetX(xKanan)
	setFont("", 10)
	pdf.MultiCell(widthKanan, lineHeight, dateStr, "", "C", false)
	pdf.Ln(2)

	yPosTtd := pdf.GetY()

	// Blok Kiri
	pdf.SetY(yPosTtd)
	pdf.SetX(xKiri)
	setFont("B", 9)
	pdf.MultiCell(widthKiri, lineHeightKop, jabatanKiri, "", "C", false)
	pdf.SetX(xKiri)
	pdf.MultiCell(widthKiri, lineHeightKop, jabatanKiri2, "", "C", false)
	
	// Blok Kanan
	pdf.SetY(yPosTtd)
	pdf.SetX(xKanan)
	setFont("B", 9)
	pdf.MultiCell(widthKanan, lineHeightKop, jabatanKanan, "", "C", false)
	pdf.SetX(xKanan)
	pdf.MultiCell(widthKanan, lineHeightKop, jabatanKanan2, "", "C", false)

	pdf.Ln(spasiTtd)
	yPosName := pdf.GetY()

	// Nama Kiri
	pdf.SetY(yPosName)
	pdf.SetX(xKiri)
	setFont("BU", 9)
	pdf.MultiCell(widthKiri, lineHeightKop, namaKiri, "", "C", false)
	setFont("", 9)
	pdf.SetX(xKiri)
	pdf.MultiCell(widthKiri, lineHeightKop, nrpKiri, "", "C", false)

	// Nama Kanan
	pdf.SetY(yPosName)
	pdf.SetX(xKanan)
	setFont("BU", 9)
	pdf.MultiCell(widthKanan, lineHeightKop, namaKanan, "", "C", false)
	setFont("", 9)
	pdf.SetX(xKanan)
	pdf.MultiCell(widthKanan, lineHeightKop, nrpKanan, "", "C", false)

	var buffer bytes.Buffer
	err = pdf.Output(&buffer)
	if err != nil { return nil, "" }

	filename := fmt.Sprintf("SKH_%s_%s.pdf", strings.ReplaceAll(doc.Resident.NamaLengkap, " ", "_"), doc.TanggalLaporan.Format("20060102"))
	return &buffer, filename
}