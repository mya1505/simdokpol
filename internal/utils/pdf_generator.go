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
	"time"

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
	
	// Dapatkan dimensi halaman dan margin terlebih dahulu
	pageWidth, _ := pdf.GetPageSize()
	marginL, _, marginR, _ := pdf.GetMargins()
	
	// Helper untuk Hitung Lebar Text Terpanjang di Blok TTD
	calcMaxW := func(lines ...string) float64 {
		setFont("B", 9)
		max := 0.0
		for _, l := range lines {
			if w := pdf.GetStringWidth(l); w > max { max = w }
		}
		return max + 2.0
	}

	// Helper untuk menambahkan garis putus-putus yang lebih presisi
	addDashedLine := func(startX, startY, endX float64) {
		pdf.SetLineWidth(0.2)
		// Pattern: 1mm dash, 0.8mm space untuk hasil yang mirip dengan contoh
		pdf.SetDashPattern([]float64{1.0, 0.8}, 0)
		pdf.Line(startX, startY, endX, startY)
		pdf.SetDashPattern([]float64{}, 0)
	}

	// Helper untuk menulis teks dengan garis putus-putus di belakangnya
	// indentFirstLine: indentasi untuk baris pertama
	// indentNextLines: indentasi untuk baris kedua dan seterusnya
	writeTextWithDash := func(indentFirstLine, indentNextLines float64, text string, font string, size float64) {
		setFont(font, size)
		
		// Hitung lebar area yang tersedia dari posisi indentasi baris pertama hingga margin kanan
		availableWidth := pageWidth - marginR - indentFirstLine
		
		// Pecah teks jika terlalu panjang
		lines := pdf.SplitLines([]byte(text), availableWidth)
		
		for i, line := range lines {
			lineText := string(line)
			
			// Gunakan indentasi yang sesuai
			currentIndent := indentFirstLine
			if i > 0 {
				currentIndent = indentNextLines
				// Recalculate available width untuk baris berikutnya
				availableWidthNext := pageWidth - marginR - indentNextLines
				// Re-split jika perlu dengan lebar baru
				if i == 1 && indentFirstLine != indentNextLines {
					remainingText := strings.Join(func() []string {
						result := make([]string, len(lines)-1)
						for j := 1; j < len(lines); j++ {
							result[j-1] = string(lines[j])
						}
						return result
					}(), " ")
					newLines := pdf.SplitLines([]byte(remainingText), availableWidthNext)
					lines = append(lines[:1], newLines...)
					lineText = string(lines[i])
				}
			}
			
			pdf.SetX(currentIndent)
			textWidth := pdf.GetStringWidth(lineText)
			pdf.Cell(textWidth, lineHeight, lineText)
			
			// Hanya tambahkan garis putus-putus di baris terakhir
			if i == len(lines)-1 {
				dashStartX := currentIndent + textWidth + 0.5
				dashEndX := pageWidth - marginR
				dashY := pdf.GetY() + (lineHeight / 2) + 0.5
				
				if dashStartX < dashEndX {
					addDashedLine(dashStartX, dashY, dashEndX)
				}
			}
			
			if i < len(lines)-1 {
				pdf.Ln(lineHeight)
			}
		}
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
	setFont("BU", 12)
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

	// BAGIAN 1: Paragraf pembuka dengan garis putus-putus (JUSTIFIED)
	setFont("", 11)
	introText := fmt.Sprintf("---- Yang bertanda tangan dibawah ini A.n. KEPALA KEPOLISIAN %s, Menerangkan dengan benar bahwa :", strings.ToUpper(config.KopBaris3))
	
	pdf.SetX(20)
	// Split manual untuk kontrol penuh
	lines := pdf.SplitLines([]byte(introText), 170)
	
	for i, line := range lines {
		lineText := string(line)
		
		if i == len(lines)-1 {
			// Baris terakhir dengan garis putus-putus, tanpa justify
			pdf.SetY(pdf.GetY())
			writeTextWithDash(20, 20, lineText, "", 11)
			pdf.Ln(lineHeight)
		} else {
			// Baris sebelumnya dengan justify
			pdf.SetX(20)
			pdf.CellFormat(170, lineHeight, lineText, "", 0, "J", false, 0, "")
			pdf.Ln(lineHeight)
		}
	}
	pdf.Ln(lineHeight)

	// Data Pemohon
	pdf.SetX(30); pdf.Cell(40, lineHeight, "Nama"); pdf.Cell(5, lineHeight, ":"); setFont("B", 11); pdf.Cell(0, lineHeight, strings.ToUpper(doc.Resident.NamaLengkap)); pdf.Ln(lineHeight)
	setFont("", 11)
	pdf.SetX(30); pdf.Cell(40, lineHeight, "TTL"); pdf.Cell(5, lineHeight, ":"); pdf.Cell(0, lineHeight, fmt.Sprintf("%s, %s", doc.Resident.TempatLahir, doc.Resident.TanggalLahir.Format("02-01-2006"))); pdf.Ln(lineHeight)
	pdf.SetX(30); pdf.Cell(40, lineHeight, "Agama"); pdf.Cell(5, lineHeight, ":"); pdf.Cell(0, lineHeight, doc.Resident.Agama); pdf.Ln(lineHeight)
	pdf.SetX(30); pdf.Cell(40, lineHeight, "Jenis kelamin"); pdf.Cell(5, lineHeight, ":"); pdf.Cell(0, lineHeight, doc.Resident.JenisKelamin); pdf.Ln(lineHeight)
	pdf.SetX(30); pdf.Cell(40, lineHeight, "Pekerjaan"); pdf.Cell(5, lineHeight, ":"); pdf.Cell(0, lineHeight, doc.Resident.Pekerjaan); pdf.Ln(lineHeight)
	pdf.SetX(30); pdf.Cell(40, lineHeight, "Alamat"); pdf.Cell(5, lineHeight, ":"); pdf.MultiCell(115, lineHeight, doc.Resident.Alamat, "", "L", false); pdf.Ln(3)

	// BAGIAN 3: Paragraf sebelum barang hilang dengan garis putus-putus (JUSTIFIED)
	reportText := fmt.Sprintf("Yang bersangkutan tersebut di atas benar telah datang di Kantor %s dan melaporkan bahwa telah kehilangan surat berharga berupa :", config.NamaKantor)
	
	lines = pdf.SplitLines([]byte(reportText), 170)
	for i, line := range lines {
		lineText := string(line)
		
		if i == len(lines)-1 {
			pdf.SetY(pdf.GetY())
			writeTextWithDash(20, 20, lineText, "", 11)
			pdf.Ln(lineHeight)
		} else {
			pdf.SetX(20)
			pdf.CellFormat(170, lineHeight, lineText, "", 0, "J", false, 0, "")
			pdf.Ln(lineHeight)
		}
	}
	pdf.Ln(3)

	// BAGIAN 2: Barang Hilang dengan format ringkas dan garis putus-putus
	for i, item := range doc.LostItems {
		pdf.SetY(pdf.GetY())
		// Format: "- 1 (Buah) NAMA_BARANG Dengan No: NOMOR A.n Pelapor"
		itemText := fmt.Sprintf("1 (Buah) %s Dengan  %s A.n Pelapor", 
			strings.ToUpper(item.NamaBarang), 
			item.Deskripsi)
		
		// Indentasi pertama di posisi 30 (untuk "- "), indentasi berikutnya di 33 (sejajar dengan teks setelah "- ")
		pdf.SetX(30)
		pdf.Cell(3, lineHeight, "-")
		writeTextWithDash(33, 33, itemText, "", 11)
		pdf.Ln(lineHeightItem)
		
		if i < len(doc.LostItems)-1 { 
			pdf.Ln(1) 
		}
	}
	pdf.Ln(2)

	// BAGIAN 4: Paragraf lokasi hilang dengan garis putus-putus (JUSTIFIED)
	lostText := fmt.Sprintf("---- Surat/kartu tersebut hilang di sekitar %s, dan sudah dilakukan pencarian namun sampai dikeluarkan Surat Keterangan ini belum ditemukan.", doc.LokasiHilang)
	
	lines = pdf.SplitLines([]byte(lostText), 170)
	for i, line := range lines {
		lineText := string(line)
		
		if i == len(lines)-1 {
			pdf.SetY(pdf.GetY())
			writeTextWithDash(20, 20, lineText, "", 11)
			pdf.Ln(lineHeight)
		} else {
			pdf.SetX(20)
			pdf.CellFormat(170, lineHeight, lineText, "", 0, "J", false, 0, "")
			pdf.Ln(lineHeight)
		}
	}
	pdf.Ln(3)

	// TTD Pemohon
	pdf.SetX(120)
	pdf.MultiCell(70, lineHeight, "Yang Bermohon", "", "C", false)
	pdf.Ln(spasiTtdBermohon)
	pdf.SetX(120)
	setFont("BU", 11)
	pdf.MultiCell(70, lineHeight, strings.ToUpper(doc.Resident.NamaLengkap), "", "C", false)
	
	// BAGIAN 5: Demikian dengan garis putus-putus (JUSTIFIED)
	setFont("", 11)
	pdf.Ln(3)
	demiText := "----- Demikian Surat Keterangan ini dibuat dengan sebenar-benarnya dan dapat dipergunakan sebagaimana perlunya."
	
	lines = pdf.SplitLines([]byte(demiText), 170)
	for i, line := range lines {
		lineText := string(line)
		
		if i == len(lines)-1 {
			pdf.SetY(pdf.GetY())
			writeTextWithDash(20, 20, lineText, "", 11)
			pdf.Ln(lineHeight)
		} else {
			pdf.SetX(20)
			pdf.CellFormat(170, lineHeight, lineText, "", 0, "J", false, 0, "")
			pdf.Ln(lineHeight)
		}
	}
	pdf.Ln(3)

	// Tindakan Yang Diambil
	setFont("BU", 11)
	pdf.SetX(20); pdf.Cell(0, lineHeight, "Tindakan Yang Diambil :"); pdf.Ln(lineHeight)
	
	// Item 1
	setFont("", 11)
	pdf.SetY(pdf.GetY())
	item1Text := "Menerima laporan dan membuat Surat Keterangan Kehilangan barang guna seperlunya;"
	pdf.SetX(25)
	pdf.Cell(5, lineHeight, "1.")
	writeTextWithDash(30, 30, item1Text, "", 11)
	pdf.Ln(lineHeightItem)
	
	// Item 2
	archiveDays := 15
	if config.ArchiveDurationDays > 0 { archiveDays = config.ArchiveDurationDays }
	pdf.SetY(pdf.GetY())
	item2Text := fmt.Sprintf("Surat keterangan kehilangan ini berlaku selama %d (%s) hari, berlaku mulai tanggal dikeluarkan;", archiveDays, IntToIndonesianWords(archiveDays))
	pdf.SetX(25)
	pdf.Cell(5, lineHeight, "2.")
	writeTextWithDash(30, 30, item2Text, "", 11)
	pdf.Ln(lineHeightItem)
	
	// Item 3 dengan garis putus-putus
	pdf.SetY(pdf.GetY())
	item3Text := "Surat Keterangan ini bukan sebagai pengganti surat yang hilang tetapi berguna untuk mengurus kembali surat yang hilang."
	pdf.SetX(25)
	pdf.Cell(5, lineHeight, "3.")
	writeTextWithDash(30, 30, item3Text, "", 11)
	pdf.Ln(lineHeightItem)
	pdf.Ln(3)

	// --- 4. TANDA TANGAN (DINAMIS WIDTH) ---
	// Helper untuk format tanggal Indonesia
	formatTanggalIndonesia := func(t time.Time) string {
		bulan := []string{
			"Januari", "Februari", "Maret", "April", "Mei", "Juni",
			"Juli", "Agustus", "September", "Oktober", "November", "Desember",
		}
		return fmt.Sprintf("%02d %s %d", t.Day(), bulan[t.Month()-1], t.Year())
	}
	
	dateStr := fmt.Sprintf("%s, %s", config.TempatSurat, formatTanggalIndonesia(doc.TanggalLaporan))
	
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

	widthKiri := calcMaxW(jabatanKiri, jabatanKiri2, namaKiri, nrpKiri)
	if widthKiri < 70 { widthKiri = 70 }
	if widthKiri > 90 { widthKiri = 90 }

	widthKanan := calcMaxW(dateStr, jabatanKanan, jabatanKanan2, namaKanan, nrpKanan)
	if widthKanan < 70 { widthKanan = 70 }
	if widthKanan > 90 { widthKanan = 90 }

	xKiri := 20.0
	xKanan := pageWidth - marginR - widthKanan

	// Render TTD
	pdf.SetX(xKanan)
	setFont("", 10)
	pdf.MultiCell(widthKanan, lineHeight, dateStr, "", "C", false)
	pdf.Ln(2)

	yPosTtd := pdf.GetY()

	pdf.SetY(yPosTtd)
	pdf.SetX(xKiri)
	setFont("B", 9)
	pdf.MultiCell(widthKiri, lineHeightKop, jabatanKiri, "", "C", false)
	pdf.SetX(xKiri)
	pdf.MultiCell(widthKiri, lineHeightKop, jabatanKiri2, "", "C", false)
	
	pdf.SetY(yPosTtd)
	pdf.SetX(xKanan)
	setFont("B", 9)
	pdf.MultiCell(widthKanan, lineHeightKop, jabatanKanan, "", "C", false)
	pdf.SetX(xKanan)
	pdf.MultiCell(widthKanan, lineHeightKop, jabatanKanan2, "", "C", false)

	pdf.Ln(spasiTtd)
	yPosName := pdf.GetY()

	pdf.SetY(yPosName)
	pdf.SetX(xKiri)
	setFont("BU", 9)
	pdf.MultiCell(widthKiri, lineHeightKop, namaKiri, "", "C", false)
	setFont("", 9)
	pdf.SetX(xKiri)
	pdf.MultiCell(widthKiri, lineHeightKop, nrpKiri, "", "C", false)

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