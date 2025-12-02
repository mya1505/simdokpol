package utils

import (
	"bytes"
	"fmt"
	"log"
	"simdokpol/internal/dto"
	"simdokpol/web"
	"strconv"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"
)

// Konstanta Layout (Disamakan dengan pdf_generator.go)
const (
	pdfMarginLeft   = 20.0
	pdfMarginTop    = 15.0
	pdfMarginRight  = 20.0
	pdfMarginBottom = 15.0
	
	lineHeightKop   = 3.8
	lineHeight      = 4.2
	
	// Posisi Logo di tengah halaman (sesuai pdf_generator.go)
	logoX           = 99.0
	logoY           = 42.0
	logoW           = 12.0
	
	// Awal konten setelah header
	contentStartY   = 56.0
)

type reportPdfGenerator struct {
	pdf    *gofpdf.Fpdf
	config *dto.AppConfig
	data   *dto.AggregateReportData
	exeDir string
	loc    *time.Location
	pageN  int
}

// GenerateAggregateReportPDF membuat laporan dengan styling yang seragam dengan surat satuan
func GenerateAggregateReportPDF(data *dto.AggregateReportData, config *dto.AppConfig, exeDir string) (*bytes.Buffer, string) {
	loc, _ := time.LoadLocation(config.ZonaWaktu)
	if loc == nil {
		loc = time.UTC
	}

	r := &reportPdfGenerator{
		pdf:    gofpdf.New("P", "mm", "A4", ""),
		config: config,
		data:   data,
		exeDir: exeDir,
		loc:    loc,
		pageN:  0,
	}

	// Setup Embed Logo
	logoBytes, err := web.Assets.ReadFile("static/img/logo.png")
	if err == nil {
		logoReader := bytes.NewReader(logoBytes)
		r.pdf.RegisterImageOptionsReader("logo_embed", gofpdf.ImageOptions{ImageType: "PNG"}, logoReader)
	} else {
		log.Printf("WARNING REPORT PDF: Gagal load logo embed: %v", err)
	}

	r.addPage()

	// Konten Laporan
	r.drawReportTitle()
	r.drawSummaryTable()
	r.drawOperatorStatsTable()
	r.drawItemCompositionTable()
	r.drawDocumentListTable()
	r.drawSignatureBlock()

	var buffer bytes.Buffer
	if err := r.pdf.Output(&buffer); err != nil {
		log.Printf("CRITICAL PDF ERROR: %v", err)
		return nil, ""
	}

	filename := fmt.Sprintf("Laporan_Agregat_%s_sd_%s.pdf", data.StartDate.Format("20060102"), data.EndDate.Format("20060102"))
	return &buffer, filename
}

// addPage menambahkan halaman dengan Header Style 'Surat Keterangan'
func (r *reportPdfGenerator) addPage() {
	r.pageN++
	r.pdf.AddPage()
	r.pdf.SetMargins(pdfMarginLeft, pdfMarginTop, pdfMarginRight)
	r.pdf.SetAutoPageBreak(true, pdfMarginBottom)

	// --- 1. KOP SURAT (Gaya Blok Kiri) ---
	kopWidth := float64(60)
	
	// Baris 1 & 2 (Kecil Bold)
	r.pdf.SetFont("Courier", "B", 9)
	r.pdf.SetXY(pdfMarginLeft, 20) // Y=20 sesuai pdf_generator
	r.pdf.MultiCell(kopWidth, lineHeightKop, r.config.KopBaris1, "", "C", false)
	
	r.pdf.SetX(pdfMarginLeft)
	r.pdf.MultiCell(kopWidth, lineHeightKop, r.config.KopBaris2, "", "C", false)
	
	// Baris 3 (Besar Bold)
	r.pdf.SetX(pdfMarginLeft)
	r.pdf.SetFont("Courier", "B", 10)
	r.pdf.MultiCell(kopWidth, lineHeightKop, r.config.KopBaris3, "", "C", false)
	
	r.pdf.Ln(2)

	// Garis Bawah KOP (Hanya sepanjang teks KOP)
	currentY := r.pdf.GetY()
	r.pdf.SetLineWidth(0.3)
	r.pdf.Line(pdfMarginLeft, currentY, pdfMarginLeft+kopWidth, currentY)
	r.pdf.SetLineWidth(0.2)

	// --- 2. LOGO (Di Tengah) ---
	// Menggunakan koordinat fix agar sama persis dengan surat satuan
	r.pdf.Image("logo_embed", logoX, logoY, logoW, 0, false, "", 0, "")

	// --- 3. SETUP AREA KONTEN ---
	// Pindahkan kursor ke bawah logo untuk mulai menulis konten
	r.pdf.SetY(contentStartY)

	// Footer Halaman
	r.pdf.SetFooterFunc(func() {
		r.pdf.SetY(-15)
		r.pdf.SetFont("Courier", "I", 8)
		r.pdf.SetTextColor(128, 128, 128)
		r.pdf.CellFormat(0, 10, fmt.Sprintf("Halaman %d", r.pageN), "", 0, "R", false, 0, "")
		r.pdf.SetTextColor(0, 0, 0) // Reset warna
	})
}

func (r *reportPdfGenerator) drawReportTitle() {
	r.pdf.SetFont("Courier", "BU", 12)
	r.pdf.CellFormat(0, 6, "LAPORAN AGREGAT DOKUMEN", "", 1, "C", false, 0, "")
	
	r.pdf.SetFont("Courier", "", 10)
	dateStr := fmt.Sprintf("Periode: %s s/d %s",
		r.data.StartDate.In(r.loc).Format("02 January 2006"),
		r.data.EndDate.In(r.loc).Format("02 January 2006"),
	)
	r.pdf.CellFormat(0, 5, dateStr, "", 1, "C", false, 0, "")
	r.pdf.Ln(8)
}

func (r *reportPdfGenerator) drawSummaryTable() {
	r.pdf.SetFont("Courier", "B", 11)
	r.pdf.Cell(0, 8, "Ringkasan Umum")
	r.pdf.Ln(8)

	r.pdf.SetFont("Courier", "", 10)
	r.pdf.SetX(pdfMarginLeft + 5)
	r.pdf.Cell(60, 6, "Total Dokumen Diterbitkan")
	r.pdf.Cell(5, 6, ":")
	r.pdf.SetFont("Courier", "B", 10)
	r.pdf.Cell(0, 6, fmt.Sprintf("%d Dokumen", r.data.TotalDocuments))
	r.pdf.Ln(10)
}

func (r *reportPdfGenerator) drawOperatorStatsTable() {
	r.pdf.SetFont("Courier", "B", 11)
	r.pdf.Cell(0, 8, "Statistik Aktivitas Operator")
	r.pdf.Ln(8)
	
	// Header
	r.pdf.SetFont("Courier", "B", 9)
	r.pdf.SetFillColor(240, 240, 240)
	r.pdf.CellFormat(15, 8, "No.", "1", 0, "C", true, 0, "")
	r.pdf.CellFormat(100, 8, "Nama Operator", "1", 0, "L", true, 0, "")
	r.pdf.CellFormat(55, 8, "Jumlah Dokumen", "1", 1, "C", true, 0, "")
	
	// Body
	r.pdf.SetFont("Courier", "", 9)
	if len(r.data.OperatorStats) == 0 {
		r.pdf.CellFormat(170, 8, "Tidak ada data.", "1", 1, "C", false, 0, "")
	} else {
		for i, stat := range r.data.OperatorStats {
			r.pdf.CellFormat(15, 7, strconv.Itoa(i+1), "1", 0, "C", false, 0, "")
			r.pdf.CellFormat(100, 7, fmt.Sprintf(" %s", stat.NamaLengkap), "1", 0, "L", false, 0, "")
			r.pdf.CellFormat(55, 7, strconv.Itoa(stat.Count), "1", 1, "C", false, 0, "")
		}
	}
	r.pdf.Ln(8)
}

func (r *reportPdfGenerator) drawItemCompositionTable() {
	r.pdf.SetFont("Courier", "B", 11)
	r.pdf.Cell(0, 8, "Statistik Barang Hilang")
	r.pdf.Ln(8)

	// Header
	r.pdf.SetFont("Courier", "B", 9)
	r.pdf.SetFillColor(240, 240, 240)
	r.pdf.CellFormat(15, 8, "No.", "1", 0, "C", true, 0, "")
	r.pdf.CellFormat(100, 8, "Jenis Barang", "1", 0, "L", true, 0, "")
	r.pdf.CellFormat(55, 8, "Jumlah Laporan", "1", 1, "C", true, 0, "")

	// Body
	r.pdf.SetFont("Courier", "", 9)
	if len(r.data.ItemComposition) == 0 {
		r.pdf.CellFormat(170, 8, "Tidak ada data.", "1", 1, "C", false, 0, "")
	} else {
		for i, stat := range r.data.ItemComposition {
			r.pdf.CellFormat(15, 7, strconv.Itoa(i+1), "1", 0, "C", false, 0, "")
			r.pdf.CellFormat(100, 7, fmt.Sprintf(" %s", stat.NamaBarang), "1", 0, "L", false, 0, "")
			r.pdf.CellFormat(55, 7, strconv.Itoa(stat.Count), "1", 1, "C", false, 0, "")
		}
	}
	r.pdf.Ln(8)
}

func (r *reportPdfGenerator) drawDocumentListTable() {
	// Paksa halaman baru untuk tabel rincian agar rapi
	r.addPage()
	
	r.pdf.SetFont("Courier", "B", 11)
	r.pdf.Cell(0, 8, "Rincian Dokumen")
	r.pdf.Ln(8)

	drawHeader := func() {
		r.pdf.SetFont("Courier", "B", 8)
		r.pdf.SetFillColor(240, 240, 240)
		r.pdf.CellFormat(10, 8, "No", "1", 0, "C", true, 0, "")
		r.pdf.CellFormat(25, 8, "Tanggal", "1", 0, "C", true, 0, "")
		r.pdf.CellFormat(50, 8, "Nomor Surat", "1", 0, "L", true, 0, "")
		r.pdf.CellFormat(45, 8, "Pemohon", "1", 0, "L", true, 0, "")
		r.pdf.CellFormat(40, 8, "Operator", "1", 1, "L", true, 0, "")
	}
	
	drawHeader()

	r.pdf.SetFont("Courier", "", 8)
	if len(r.data.DocumentList) == 0 {
		r.pdf.CellFormat(170, 8, "Tidak ada dokumen.", "1", 1, "C", false, 0, "")
	} else {
		for i, doc := range r.data.DocumentList {
			// Cek sisa halaman, jika kurang buat halaman baru
			if r.pdf.GetY() > 270 {
				r.addPage()
				r.pdf.Ln(5)
				drawHeader()
				r.pdf.SetFont("Courier", "", 8)
			}
			
			r.pdf.CellFormat(10, 6, strconv.Itoa(i+1), "1", 0, "C", false, 0, "")
			r.pdf.CellFormat(25, 6, doc.TanggalLaporan.In(r.loc).Format("02-01-2006"), "1", 0, "C", false, 0, "")
			r.pdf.CellFormat(50, 6, fmt.Sprintf(" %s", doc.NomorSurat), "1", 0, "L", false, 0, "")
			
			// Truncate nama jika terlalu panjang
			pemohon := doc.Resident.NamaLengkap
			if len(pemohon) > 22 { pemohon = pemohon[:22] + "..." }
			r.pdf.CellFormat(45, 6, fmt.Sprintf(" %s", pemohon), "1", 0, "L", false, 0, "")
			
			operator := doc.Operator.NamaLengkap
			if len(operator) > 18 { operator = operator[:18] + "..." }
			r.pdf.CellFormat(40, 6, fmt.Sprintf(" %s", operator), "1", 1, "L", false, 0, "")
		}
	}
	r.pdf.Ln(8)
}

func (r *reportPdfGenerator) drawSignatureBlock() {
	// Cek sisa halaman
	if r.pdf.GetY() > 240 {
		r.addPage()
	}

	r.pdf.SetFont("Courier", "", 10)
	
	// Posisi kanan untuk TTD
	xPos := 120.0
	
	r.pdf.SetX(xPos)
	r.pdf.MultiCell(70, 5, fmt.Sprintf("%s, %s", r.config.TempatSurat, time.Now().In(r.loc).Format("02 January 2006")), "", "C", false)
	r.pdf.Ln(1)

	r.pdf.SetX(xPos)
	r.pdf.SetFont("Courier", "B", 10)
	r.pdf.MultiCell(70, 5, fmt.Sprintf("a.n. KEPALA KEPOLISIAN %s", strings.ToUpper(r.config.KopBaris3)), "", "C", false)
	
	r.pdf.SetX(xPos)
	r.pdf.MultiCell(70, 5, "KANIT SPKT", "", "C", false)
	
	r.pdf.Ln(20) // Ruang Tanda Tangan

	r.pdf.SetX(xPos)
	r.pdf.SetFont("Courier", "BU", 10)
	r.pdf.MultiCell(70, 5, "( .............................. )", "", "C", false)
	
	r.pdf.SetX(xPos)
	r.pdf.SetFont("Courier", "", 10)
	r.pdf.MultiCell(70, 5, "NRP. .........................", "", "C", false)
}