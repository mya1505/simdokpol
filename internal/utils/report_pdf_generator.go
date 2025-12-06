package utils

import (
	"bytes"
	"fmt"
	"log"
	"math"
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
	
	// --- FITUR BARU: GRAFIK ---
	r.drawItemBarChart() // <--- Sisipkan Grafik di sini
	// --------------------------

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

	// KOP Header
	r.pdf.SetFont("Courier", "B", 10)
	w1 := r.pdf.GetStringWidth(r.config.KopBaris1)
	w2 := r.pdf.GetStringWidth(r.config.KopBaris2)
	w3 := r.pdf.GetStringWidth(r.config.KopBaris3)
	maxTextWidth := math.Max(w1, math.Max(w2, w3))
	
	kopWidth := maxTextWidth + 2.0
	if kopWidth < 60.0 { kopWidth = 60.0 }
	if kopWidth > 75.0 { kopWidth = 75.0 }

	r.pdf.SetFont("Courier", "B", 9)
	r.pdf.SetXY(pdfMarginLeft, 20)
	r.pdf.MultiCell(kopWidth, lineHeightKop, r.config.KopBaris1, "", "C", false)
	
	r.pdf.SetX(pdfMarginLeft)
	r.pdf.MultiCell(kopWidth, lineHeightKop, r.config.KopBaris2, "", "C", false)
	
	r.pdf.SetX(pdfMarginLeft)
	r.pdf.SetFont("Courier", "B", 10)
	r.pdf.MultiCell(kopWidth, lineHeightKop, r.config.KopBaris3, "", "C", false)
	
	r.pdf.Ln(2)
	currentY := r.pdf.GetY()
	r.pdf.SetLineWidth(0.3)
	r.pdf.Line(pdfMarginLeft, currentY, pdfMarginLeft+kopWidth, currentY)
	r.pdf.SetLineWidth(0.2)

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

// --- FITUR BARU: FUNGSI GAMBAR GRAFIK BATANG ---
func (r *reportPdfGenerator) drawItemBarChart() {
	// Cek sisa halaman, jika sempit buat halaman baru
	if r.pdf.GetY() > 200 {
		r.addPage()
	}
	
	r.pdf.SetFont("Courier", "B", 11)
	r.pdf.Cell(0, 8, "Grafik Statistik Barang Hilang")
	r.pdf.Ln(8)

	if len(r.data.ItemComposition) == 0 {
		r.pdf.SetFont("Courier", "I", 10)
		r.pdf.Cell(0, 8, "(Tidak ada data untuk ditampilkan)")
		r.pdf.Ln(10)
		return
	}

	// 1. Cari Nilai Tertinggi untuk Skala
	maxVal := 0
	for _, item := range r.data.ItemComposition {
		if item.Count > maxVal {
			maxVal = item.Count
		}
	}
	if maxVal == 0 { maxVal = 1 } // Prevent divide by zero

	// Konfigurasi Grafik
	barHeight := 6.0
	gap := 4.0
	labelWidth := 40.0 // Lebar area label kiri
	chartMaxWidth := 120.0 // Lebar maksimal batang
	
	r.pdf.SetFont("Courier", "", 9)

	for _, item := range r.data.ItemComposition {
		// Cek halaman
		if r.pdf.GetY() > 260 {
			r.addPage()
			r.pdf.Ln(10)
		}

		// Label Kiri
		r.pdf.CellFormat(labelWidth, barHeight, item.NamaBarang, "", 0, "R", false, 0, "")
		
		// Hitung Panjang Batang
		barWidth := (float64(item.Count) / float64(maxVal)) * chartMaxWidth
		if barWidth < 1 { barWidth = 1 } // Minimal terlihat dikit
		
		// Gambar Batang (Warna Biru Polisi #4e73df)
		r.pdf.SetFillColor(78, 115, 223) 
		xPos := r.pdf.GetX() + 2
		yPos := r.pdf.GetY()
		r.pdf.Rect(xPos, yPos, barWidth, barHeight, "F")
		
		// Tulis Angka di Ujung Batang
		r.pdf.SetX(xPos + barWidth + 2)
		r.pdf.SetTextColor(100, 100, 100)
		r.pdf.Cell(20, barHeight, strconv.Itoa(item.Count))
		r.pdf.SetTextColor(0, 0, 0) // Reset warna

		r.pdf.Ln(barHeight + gap)
	}
	
	r.pdf.Ln(5) // Spasi bawah
}
// ----------------------------------------------------

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
	r.pdf.Cell(0, 8, "Rincian Barang Hilang")
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
	r.pdf.Cell(0, 8, "Daftar Lengkap Dokumen Diterbitkan")
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