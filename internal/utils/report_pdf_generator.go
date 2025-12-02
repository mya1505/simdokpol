package utils

import (
	"bytes"
	"fmt"
	"log"
	"math" // <-- Tambah import Math
	"simdokpol/internal/dto"
	"simdokpol/web"
	"strconv"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"
)

const (
	pdfMarginLeft   = 20.0
	pdfMarginTop    = 15.0
	pdfMarginRight  = 20.0
	pdfMarginBottom = 15.0
	
	lineHeightKop   = 3.8
	lineHeight      = 4.2
	
	logoX           = 99.0
	logoY           = 42.0
	logoW           = 12.0
	
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

	// Load Logo
	logoBytes, err := web.Assets.ReadFile("static/img/logo.png")
	if err == nil {
		logoReader := bytes.NewReader(logoBytes)
		r.pdf.RegisterImageOptionsReader("logo_embed", gofpdf.ImageOptions{ImageType: "PNG"}, logoReader)
	} else {
		log.Printf("WARNING REPORT PDF: Gagal load logo embed: %v", err)
	}

	r.addPage()

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

func (r *reportPdfGenerator) addPage() {
	r.pageN++
	r.pdf.AddPage()
	r.pdf.SetMargins(pdfMarginLeft, pdfMarginTop, pdfMarginRight)
	r.pdf.SetAutoPageBreak(true, pdfMarginBottom)

	// --- LOGIKA DINAMIS LEBAR KOP ---
	// 1. Set font terbesar yang digunakan di KOP untuk pengukuran (Bold 10)
	r.pdf.SetFont("Courier", "B", 10)

	// 2. Ukur panjang masing-masing baris
	w1 := r.pdf.GetStringWidth(r.config.KopBaris1)
	w2 := r.pdf.GetStringWidth(r.config.KopBaris2)
	w3 := r.pdf.GetStringWidth(r.config.KopBaris3)

	// 3. Ambil yang paling panjang
	maxWidth := math.Max(w1, math.Max(w2, w3))

	// 4. Tambahkan padding dan Batasan
	// Minimal 60mm, Maksimal 75mm (agar tidak menabrak logo di X=99)
	kopWidth := maxWidth + 2.0 // Padding 2mm
	
	if kopWidth < 60.0 {
		kopWidth = 60.0
	}
	if kopWidth > 75.0 {
		kopWidth = 75.0 // Cap di 75mm agar tidak overlap logo
	}
	// --------------------------------

	// Baris 1 & 2 (Kecil Bold)
	r.pdf.SetFont("Courier", "B", 9)
	r.pdf.SetXY(pdfMarginLeft, 20)
	r.pdf.MultiCell(kopWidth, lineHeightKop, r.config.KopBaris1, "", "C", false)
	
	r.pdf.SetX(pdfMarginLeft)
	r.pdf.MultiCell(kopWidth, lineHeightKop, r.config.KopBaris2, "", "C", false)
	
	// Baris 3 (Besar Bold)
	r.pdf.SetX(pdfMarginLeft)
	r.pdf.SetFont("Courier", "B", 10)
	r.pdf.MultiCell(kopWidth, lineHeightKop, r.config.KopBaris3, "", "C", false)
	
	r.pdf.Ln(2)

	// Garis Bawah KOP (Dinamis mengikuti lebar kopWidth)
	currentY := r.pdf.GetY()
	r.pdf.SetLineWidth(0.3)
	r.pdf.Line(pdfMarginLeft, currentY, pdfMarginLeft+kopWidth, currentY)
	r.pdf.SetLineWidth(0.2)

	// Logo (Tetap di tengah)
	r.pdf.Image("logo_embed", logoX, logoY, logoW, 0, false, "", 0, "")

	r.pdf.SetY(contentStartY)

	r.pdf.SetFooterFunc(func() {
		r.pdf.SetY(-15)
		r.pdf.SetFont("Courier", "I", 8)
		r.pdf.SetTextColor(128, 128, 128)
		r.pdf.CellFormat(0, 10, fmt.Sprintf("Halaman %d", r.pageN), "", 0, "R", false, 0, "")
		r.pdf.SetTextColor(0, 0, 0)
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
	
	r.pdf.SetFont("Courier", "B", 9)
	r.pdf.SetFillColor(240, 240, 240)
	r.pdf.CellFormat(15, 8, "No.", "1", 0, "C", true, 0, "")
	r.pdf.CellFormat(100, 8, "Nama Operator", "1", 0, "L", true, 0, "")
	r.pdf.CellFormat(55, 8, "Jumlah Dokumen", "1", 1, "C", true, 0, "")
	
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

	r.pdf.SetFont("Courier", "B", 9)
	r.pdf.SetFillColor(240, 240, 240)
	r.pdf.CellFormat(15, 8, "No.", "1", 0, "C", true, 0, "")
	r.pdf.CellFormat(100, 8, "Jenis Barang", "1", 0, "L", true, 0, "")
	r.pdf.CellFormat(55, 8, "Jumlah Laporan", "1", 1, "C", true, 0, "")

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
			if r.pdf.GetY() > 270 {
				r.addPage()
				r.pdf.Ln(5)
				drawHeader()
				r.pdf.SetFont("Courier", "", 8)
			}
			
			r.pdf.CellFormat(10, 6, strconv.Itoa(i+1), "1", 0, "C", false, 0, "")
			r.pdf.CellFormat(25, 6, doc.TanggalLaporan.In(r.loc).Format("02-01-2006"), "1", 0, "C", false, 0, "")
			r.pdf.CellFormat(50, 6, fmt.Sprintf(" %s", doc.NomorSurat), "1", 0, "L", false, 0, "")
			
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
	if r.pdf.GetY() > 240 {
		r.addPage()
	}

	r.pdf.SetFont("Courier", "", 10)
	xPos := 120.0
	
	r.pdf.SetX(xPos)
	r.pdf.MultiCell(70, 5, fmt.Sprintf("%s, %s", r.config.TempatSurat, time.Now().In(r.loc).Format("02 January 2006")), "", "C", false)
	r.pdf.Ln(1)

	r.pdf.SetX(xPos)
	r.pdf.SetFont("Courier", "B", 10)
	r.pdf.MultiCell(70, 5, fmt.Sprintf("a.n. KEPALA KEPOLISIAN %s", strings.ToUpper(r.config.KopBaris3)), "", "C", false)
	
	r.pdf.SetX(xPos)
	r.pdf.MultiCell(70, 5, "KANIT SPKT", "", "C", false) 
	
	r.pdf.Ln(20)

	r.pdf.SetX(xPos)
	r.pdf.SetFont("Courier", "BU", 10)
	r.pdf.MultiCell(70, 5, "( .............................. )", "", "C", false)
	
	r.pdf.SetX(xPos)
	r.pdf.SetFont("Courier", "", 10)
	r.pdf.MultiCell(70, 5, "NRP. .........................", "", "C", false)
}