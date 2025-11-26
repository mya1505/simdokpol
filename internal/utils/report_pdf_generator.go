package utils

import (
	"bytes"
	"fmt"
	"path/filepath"
	"simdokpol/internal/dto"
	// "simdokpol/internal/models" // <-- PERBAIKAN: HAPUS BARIS INI
	"strconv"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"
)

// Definisikan konstanta untuk PDF
const (
	pdfMarginLeft   = 15.0
	pdfMarginTop    = 15.0
	pdfMarginRight  = 15.0
	pdfMarginBottom = 20.0
	pdfLineHeight   = 5.0
)

// reportPdfGenerator adalah struct helper untuk mengelola state PDF
type reportPdfGenerator struct {
	pdf    *gofpdf.Fpdf
	config *dto.AppConfig
	data   *dto.AggregateReportData
	exeDir string
	loc    *time.Location
	pageN  int
}

// GenerateAggregateReportPDF adalah fungsi publik utama
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

	r.addPage() // Tambah halaman pertama

	// Mulai menggambar konten
	r.drawReportTitle()
	r.drawSummaryTable()
	r.drawOperatorStatsTable()
	r.drawItemCompositionTable()
	r.drawDocumentListTable()
	r.drawSignatureBlock()

	// Finalisasi
	var buffer bytes.Buffer
	err := r.pdf.Output(&buffer)
	if err != nil {
		// Jika terjadi error, kembalikan buffer error
		r.pdf.AddPage()
		r.pdf.SetFont("Courier", "B", 16)
		r.pdf.Cell(0, 10, "Gagal membuat PDF: "+err.Error())
		r.pdf.Output(&buffer)
	}

	filename := fmt.Sprintf("Laporan_Agregat_%s_sd_%s.pdf", data.StartDate.Format("20060102"), data.EndDate.Format("20060102"))
	return &buffer, filename
}

// addPage adalah helper untuk menambah halaman baru (dengan KOP dan footer)
func (r *reportPdfGenerator) addPage() {
	r.pageN++
	r.pdf.AddPage()
	r.pdf.SetMargins(pdfMarginLeft, pdfMarginTop, pdfMarginRight)
	r.pdf.SetAutoPageBreak(true, pdfMarginBottom)

	// --- 1. KOP SURAT ---
	logoPath := filepath.Join(r.exeDir, "web", "static", "img", "logo.png")
	r.pdf.Image(logoPath, pdfMarginLeft, pdfMarginTop, 15, 0, false, "", 0, "")
	
	r.pdf.SetFont("Courier", "B", 10)
	kopWidth, _ := r.pdf.GetPageSize()
	kopWidth -= (pdfMarginLeft + pdfMarginRight + 20) // 20 utk spasi logo
	
	r.pdf.SetX(pdfMarginLeft + 20)
	r.pdf.MultiCell(kopWidth, 4, r.config.KopBaris1, "", "C", false)
	r.pdf.SetX(pdfMarginLeft + 20)
	r.pdf.MultiCell(kopWidth, 4, r.config.KopBaris2, "", "C", false)
	r.pdf.SetX(pdfMarginLeft + 20)
	r.pdf.SetFont("Courier", "B", 11)
	r.pdf.MultiCell(kopWidth, 4, r.config.KopBaris3, "", "C", false)

	// Garis
	yPos := r.pdf.GetY()
	if yPos < pdfMarginTop+15 { yPos = pdfMarginTop + 15 } // Pastikan Y pos setelah logo
	r.pdf.SetLineWidth(0.5)
	r.pdf.Line(pdfMarginLeft, yPos+1, 210-pdfMarginRight, yPos+1)
	r.pdf.SetLineWidth(0.2)
	
	r.pdf.SetY(yPos + 3) // Pindah ke bawah garis

	// --- 2. FOOTER HALAMAN ---
	r.pdf.SetFooterFunc(func() {
		r.pdf.SetY(-pdfMarginBottom + 5) // 15mm dari bawah
		r.pdf.SetFont("Courier", "I", 8)
		r.pdf.SetTextColor(128, 128, 128)
		// Teks Kiri
		r.pdf.CellFormat(0, 10, fmt.Sprintf("SIMDOKPOL - Laporan Agregat v%s", r.config.KopBaris3), "", 0, "L", false, 0, "")
		// Teks Kanan (Nomor Halaman)
		r.pdf.CellFormat(0, 10, fmt.Sprintf("Halaman %d", r.pageN), "", 0, "R", false, 0, "")
	})
}

// drawReportTitle menggambar judul laporan
func (r *reportPdfGenerator) drawReportTitle() {
	r.pdf.SetFont("Courier", "BU", 12)
	r.pdf.CellFormat(0, pdfLineHeight*2, "LAPORAN AGREGAT DOKUMEN", "", 1, "C", false, 0, "")
	
	r.pdf.SetFont("Courier", "", 10)
	dateStr := fmt.Sprintf("Periode: %s s/d %s",
		r.data.StartDate.In(r.loc).Format("02 January 2006"),
		r.data.EndDate.In(r.loc).Format("02 January 2006"),
	)
	r.pdf.CellFormat(0, pdfLineHeight, dateStr, "", 1, "C", false, 0, "")
	r.pdf.Ln(pdfLineHeight)
}

// drawSummaryTable menggambar tabel ringkasan
func (r *reportPdfGenerator) drawSummaryTable() {
	r.pdf.SetFont("Courier", "B", 11)
	r.pdf.Cell(0, pdfLineHeight*1.5, "Ringkasan Umum")
	r.pdf.Ln(pdfLineHeight * 1.5)

	r.pdf.SetFont("Courier", "", 10)
	r.pdf.SetX(pdfMarginLeft + 5)
	r.pdf.Cell(80, pdfLineHeight, "Total Dokumen Diterbitkan")
	r.pdf.Cell(10, pdfLineHeight, ":")
	r.pdf.SetFont("Courier", "B", 10)
	r.pdf.Cell(0, pdfLineHeight, fmt.Sprintf("%d Dokumen", r.data.TotalDocuments))
	r.pdf.Ln(pdfLineHeight * 1.5)
}

// drawOperatorStatsTable menggambar tabel statistik operator
func (r *reportPdfGenerator) drawOperatorStatsTable() {
	r.pdf.SetFont("Courier", "B", 11)
	r.pdf.Cell(0, pdfLineHeight*1.5, "Statistik Aktivitas Operator")
	r.pdf.Ln(pdfLineHeight * 1.5)
	
	// Header Tabel
	r.pdf.SetFont("Courier", "B", 9)
	r.pdf.SetFillColor(230, 230, 230)
	r.pdf.CellFormat(20, pdfLineHeight*1.5, "No.", "1", 0, "C", true, 0, "")
	r.pdf.CellFormat(100, pdfLineHeight*1.5, "Nama Operator", "1", 0, "L", true, 0, "")
	r.pdf.CellFormat(60, pdfLineHeight*1.5, "Jumlah Dokumen", "1", 1, "C", true, 0, "")
	
	// Body Tabel
	r.pdf.SetFont("Courier", "", 9)
	if len(r.data.OperatorStats) == 0 {
		r.pdf.CellFormat(180, pdfLineHeight*1.5, "Tidak ada aktivitas operator pada rentang tanggal ini.", "1", 1, "C", false, 0, "")
	} else {
		for i, stat := range r.data.OperatorStats {
			r.pdf.CellFormat(20, pdfLineHeight*1.5, strconv.Itoa(i+1), "1", 0, "C", false, 0, "")
			r.pdf.CellFormat(100, pdfLineHeight*1.5, fmt.Sprintf(" %s", stat.NamaLengkap), "1", 0, "L", false, 0, "")
			r.pdf.CellFormat(60, pdfLineHeight*1.5, strconv.Itoa(stat.Count), "1", 1, "C", false, 0, "")
		}
	}
	r.pdf.Ln(pdfLineHeight)
}

// drawItemCompositionTable menggambar tabel komposisi barang
func (r *reportPdfGenerator) drawItemCompositionTable() {
	r.pdf.SetFont("Courier", "B", 11)
	r.pdf.Cell(0, pdfLineHeight*1.5, "Statistik Komposisi Barang Hilang")
	r.pdf.Ln(pdfLineHeight * 1.5)

	// Header Tabel
	r.pdf.SetFont("Courier", "B", 9)
	r.pdf.SetFillColor(230, 230, 230)
	r.pdf.CellFormat(20, pdfLineHeight*1.5, "No.", "1", 0, "C", true, 0, "")
	r.pdf.CellFormat(100, pdfLineHeight*1.5, "Jenis Barang", "1", 0, "L", true, 0, "")
	r.pdf.CellFormat(60, pdfLineHeight*1.5, "Jumlah Laporan", "1", 1, "C", true, 0, "")

	// Body Tabel
	r.pdf.SetFont("Courier", "", 9)
	if len(r.data.ItemComposition) == 0 {
		r.pdf.CellFormat(180, pdfLineHeight*1.5, "Tidak ada laporan barang hilang pada rentang tanggal ini.", "1", 1, "C", false, 0, "")
	} else {
		for i, stat := range r.data.ItemComposition {
			r.pdf.CellFormat(20, pdfLineHeight*1.5, strconv.Itoa(i+1), "1", 0, "C", false, 0, "")
			r.pdf.CellFormat(100, pdfLineHeight*1.5, fmt.Sprintf(" %s", stat.NamaBarang), "1", 0, "L", false, 0, "")
			r.pdf.CellFormat(60, pdfLineHeight*1.5, strconv.Itoa(stat.Count), "1", 1, "C", false, 0, "")
		}
	}
	r.pdf.Ln(pdfLineHeight)
}

// drawDocumentListTable menggambar tabel daftar dokumen (bisa multi-halaman)
func (r *reportPdfGenerator) drawDocumentListTable() {
	r.addPage() // Mulai daftar dokumen di halaman baru
	r.pdf.SetFont("Courier", "B", 11)
	r.pdf.Cell(0, pdfLineHeight*1.5, "Daftar Lengkap Dokumen Diterbitkan")
	r.pdf.Ln(pdfLineHeight * 1.5)

	// Header Tabel
	drawHeader := func() {
		r.pdf.SetFont("Courier", "B", 8)
		r.pdf.SetFillColor(230, 230, 230)
		r.pdf.CellFormat(10, pdfLineHeight*1.5, "No", "1", 0, "C", true, 0, "")
		r.pdf.CellFormat(25, pdfLineHeight*1.5, "Tgl Laporan", "1", 0, "C", true, 0, "")
		r.pdf.CellFormat(55, pdfLineHeight*1.5, "Nomor Surat", "1", 0, "L", true, 0, "")
		r.pdf.CellFormat(45, pdfLineHeight*1.5, "Pemohon", "1", 0, "L", true, 0, "")
		r.pdf.CellFormat(45, pdfLineHeight*1.5, "Operator", "1", 1, "L", true, 0, "")
	}
	
	drawHeader()

	// Body Tabel
	r.pdf.SetFont("Courier", "", 8)
	if len(r.data.DocumentList) == 0 {
		r.pdf.CellFormat(180, pdfLineHeight*1.5, "Tidak ada dokumen diterbitkan pada rentang tanggal ini.", "1", 1, "C", false, 0, "")
	} else {
		for i, doc := range r.data.DocumentList {
			// Cek jika butuh halaman baru
			if r.pdf.GetY()+pdfLineHeight*1.5 > (297 - pdfMarginBottom) {
				r.addPage()
				r.pdf.Ln(pdfLineHeight) // Spasi dari KOP
				drawHeader()
				r.pdf.SetFont("Courier", "", 8)
			}
			
			r.pdf.CellFormat(10, pdfLineHeight*1.5, strconv.Itoa(i+1), "1", 0, "C", false, 0, "")
			r.pdf.CellFormat(25, pdfLineHeight*1.5, doc.TanggalLaporan.In(r.loc).Format("02-01-2006"), "1", 0, "C", false, 0, "")
			r.pdf.CellFormat(55, pdfLineHeight*1.5, fmt.Sprintf(" %s", doc.NomorSurat), "1", 0, "L", false, 0, "")
			r.pdf.CellFormat(45, pdfLineHeight*1.5, fmt.Sprintf(" %s", doc.Resident.NamaLengkap), "1", 0, "L", false, 0, "")
			r.pdf.CellFormat(45, pdfLineHeight*1.5, fmt.Sprintf(" %s", doc.Operator.NamaLengkap), "1", 1, "L", false, 0, "")
		}
	}
	r.pdf.Ln(pdfLineHeight)
}

// drawSignatureBlock menggambar blok tanda tangan
func (r *reportPdfGenerator) drawSignatureBlock() {
	// Cek jika butuh halaman baru
	if r.pdf.GetY()+50 > (297 - pdfMarginBottom) {
		r.addPage()
		r.pdf.Ln(pdfLineHeight * 2)
	}

	r.pdf.SetFont("Courier", "", 10)
	
	// Tanggal
	r.pdf.SetX(120)
	r.pdf.MultiCell(70, pdfLineHeight, fmt.Sprintf("%s, %s", r.config.TempatSurat, time.Now().In(r.loc).Format("02 January 2006")), "", "C", false)
	r.pdf.Ln(2)

	// Jabatan
	r.pdf.SetX(120)
	r.pdf.SetFont("Courier", "B", 10)
	r.pdf.MultiCell(70, pdfLineHeight, fmt.Sprintf("a.n. KEPALA KEPOLISIAN %s", strings.ToUpper(r.config.KopBaris3)), "", "C", false)
	r.pdf.SetX(120)
	r.pdf.MultiCell(70, pdfLineHeight, "KANIT SPKT I", "", "C", false) // Asumsi
	
	r.pdf.Ln(pdfLineHeight * 3) // Spasi tanda tangan

	// Nama & NRP
	r.pdf.SetX(120)
	r.pdf.SetFont("Courier", "BU", 10)
	r.pdf.MultiCell(70, pdfLineHeight, "NAMA PEJABAT (DEFAULT)", "", "C", false) // TODO: Buat dinamis jika perlu
	r.pdf.SetX(120)
	r.pdf.SetFont("Courier", "", 10)
	r.pdf.MultiCell(70, pdfLineHeight, "PANGKAT / NRP. 12345678", "", "C", false)
}