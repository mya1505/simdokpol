package services

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"simdokpol/internal/dto"
	"simdokpol/internal/models"
	"simdokpol/internal/repositories"
	"simdokpol/internal/utils"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type LostDocumentService interface {
	CreateLostDocument(residentData models.Resident, items []models.LostItem, operatorID uint, lokasiHilang string, petugasPelaporID uint, pejabatPersetujuID uint) (*models.LostDocument, error)
	UpdateLostDocument(docID uint, residentData models.Resident, items []models.LostItem, lokasiHilang string, petugasPelaporID uint, pejabatPersetujuID uint, loggedInUserID uint) (*models.LostDocument, error)
	FindAll(query string, statusFilter string) ([]models.LostDocument, error)
	SearchGlobal(query string) ([]models.LostDocument, error)
	FindByID(id uint, actorID uint) (*models.LostDocument, error)
	DeleteLostDocument(id uint, loggedInUserID uint) error
	ExportDocuments(query string, statusFilter string) (*bytes.Buffer, string, error)
	GenerateDocumentPDF(docID uint, actorID uint) (*bytes.Buffer, string, error)
	GetDocumentsPaged(req dto.DataTableRequest, statusFilter string) (*dto.DataTableResponse, error)
}

type lostDocumentService struct {
	db           *gorm.DB
	docRepo      repositories.LostDocumentRepository
	residentRepo repositories.ResidentRepository
	userRepo     repositories.UserRepository
	auditService AuditLogService
	configService ConfigService
	configRepo   repositories.ConfigRepository
	exeDir       string
	docNumMutex  sync.Mutex
}

func NewLostDocumentService(db *gorm.DB, docRepo repositories.LostDocumentRepository, residentRepo repositories.ResidentRepository, userRepo repositories.UserRepository, auditService AuditLogService, configService ConfigService, configRepo repositories.ConfigRepository, exeDir string) LostDocumentService {
	return &lostDocumentService{
		db:            db,
		docRepo:       docRepo,
		residentRepo:  residentRepo,
		userRepo:      userRepo,
		auditService:  auditService,
		configService: configService,
		configRepo:    configRepo,
		exeDir:        exeDir,
	}
}

func (s *lostDocumentService) GetDocumentsPaged(req dto.DataTableRequest, statusFilter string) (*dto.DataTableResponse, error) {
	appConfig, err := s.configService.GetConfig()
	if err != nil {
		return nil, err
	}
	archiveDays := 15
	if appConfig != nil && appConfig.ArchiveDurationDays > 0 {
		archiveDays = appConfig.ArchiveDurationDays
	}

	docs, total, filtered, err := s.docRepo.FindAllPaged(req, statusFilter, archiveDays)
	if err != nil {
		return nil, err
	}

	docs, _ = s.processDocsStatus(docs)

	return &dto.DataTableResponse{
		Draw:            req.Draw,
		RecordsTotal:    total,
		RecordsFiltered: filtered,
		Data:            docs,
	}, nil
}

func (s *lostDocumentService) generateTempNIK() string {
	timestampNano := time.Now().UnixNano()
	timestampStr := fmt.Sprintf("%d", timestampNano)
	if len(timestampStr) > 12 {
		timestampStr = timestampStr[len(timestampStr)-12:]
	}
	return fmt.Sprintf("TEMP%s", timestampStr)
}

func (s *lostDocumentService) GenerateDocumentPDF(docID uint, actorID uint) (*bytes.Buffer, string, error) {
	doc, err := s.FindByID(docID, actorID)
	if err != nil {
		return nil, "", err
	}

	config, err := s.configService.GetConfig()
	if err != nil {
		return nil, "", fmt.Errorf("gagal memuat konfigurasi: %w", err)
	}

	buffer, filename := utils.GenerateLostDocumentPDF(doc, config, s.exeDir)
	return buffer, filename, nil
}

func (s *lostDocumentService) ExportDocuments(query string, statusFilter string) (*bytes.Buffer, string, error) {
	docs, err := s.FindAll(query, statusFilter)
	if err != nil {
		return nil, "", err
	}

	f := excelize.NewFile()
	sheet := "Data Dokumen"
	if _, err := f.NewSheet(sheet); err != nil {
		return nil, "", err
	}
	f.DeleteSheet("Sheet1")
	
	loc, _ := s.configService.GetLocation()
	if loc == nil {
		loc = time.UTC
	}

	headers := []string{
		"No.", "Nomor Surat", "Tanggal Laporan", "Status", "Nama Pemohon", "NIK", "TTL", "Alamat",
		"Barang Hilang (Nama)", "Barang Hilang (Deskripsi)", "Lokasi Hilang", "Operator (Pembuat)",
		"Petugas Pelapor", "Pejabat Persetuju",
	}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, header)
	}

	for i, doc := range docs {
		row := i + 2 
		var itemNames, itemDescs []string
		for _, item := range doc.LostItems {
			itemNames = append(itemNames, item.NamaBarang)
			itemDescs = append(itemDescs, item.Deskripsi)
		}

		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), i+1)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), doc.NomorSurat)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), doc.TanggalLaporan.In(loc).Format("02-01-2006 15:04"))
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), doc.Status)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), doc.Resident.NamaLengkap)
		f.SetCellValue(sheet, fmt.Sprintf("F%d", row), doc.Resident.NIK)
		f.SetCellValue(sheet, fmt.Sprintf("G%d", row), fmt.Sprintf("%s, %s", doc.Resident.TempatLahir, doc.Resident.TanggalLahir.Format("02-01-2006")))
		f.SetCellValue(sheet, fmt.Sprintf("H%d", row), doc.Resident.Alamat)
		f.SetCellValue(sheet, fmt.Sprintf("I%d", row), strings.Join(itemNames, ", "))
		f.SetCellValue(sheet, fmt.Sprintf("J%d", row), strings.Join(itemDescs, ", "))
		f.SetCellValue(sheet, fmt.Sprintf("K%d", row), doc.LokasiHilang)
		f.SetCellValue(sheet, fmt.Sprintf("L%d", row), doc.Operator.NamaLengkap)
		f.SetCellValue(sheet, fmt.Sprintf("M%d", row), doc.PetugasPelapor.NamaLengkap)
		f.SetCellValue(sheet, fmt.Sprintf("N%d", row), doc.PejabatPersetuju.NamaLengkap)
	}

	buffer, err := f.WriteToBuffer()
	if err != nil {
		return nil, "", err
	}

	filename := fmt.Sprintf("Export_Dokumen_%s_%s.xlsx", statusFilter, time.Now().In(loc).Format("20060102_150405"))
	return buffer, filename, nil
}

func (s *lostDocumentService) FindByID(id uint, actorID uint) (*models.LostDocument, error) {
	doc, err := s.docRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	actor, err := s.userRepo.FindByID(actorID)
	if err != nil {
		return nil, errors.New("pengguna tidak valid")
	}

	if actor.Peran != models.RoleSuperAdmin && doc.OperatorID != actorID {
		return nil, ErrAccessDenied
	}

	appConfig, _ := s.configService.GetConfig()
	
	// --- FIX BUG (SA5011): Cek Nil sebelum akses ---
	if appConfig != nil && appConfig.ArchiveDurationDays > 0 {
		archiveDuration := time.Duration(appConfig.ArchiveDurationDays) * 24 * time.Hour
		if doc.Status == "DITERBITKAN" && time.Now().After(doc.TanggalLaporan.Add(archiveDuration)) {
			doc.Status = "DIARSIPKAN"
		}
	}
	// --- END FIX ---

	return doc, nil
}

func (s *lostDocumentService) generateDocumentNumber(tx *gorm.DB) (string, string, error) {
	loc, err := s.configService.GetLocation()
	if err != nil {
		loc = time.UTC
	}
	now := time.Now().In(loc)
	year := now.Year()
	month := int(now.Month())
	monthRoman := intToRoman(month)

	appConfig, err := s.configService.GetConfig()
	if err != nil {
		return "", "", fmt.Errorf("gagal memuat konfigurasi: %w", err)
	}
	// Safety check jika config nil (fallback default format)
	format := "SKH/%03d/%s/TUK.7.2.1/%d"
	if appConfig != nil && appConfig.FormatNomorSurat != "" {
		format = appConfig.FormatNomorSurat
	}

	lastNumFromDB := 0
	lastDoc, err := s.docRepo.GetLastDocumentOfYear(tx, year)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", "", fmt.Errorf("gagal query dokumen terakhir: %w", err)
	}
	if err == nil && lastDoc != nil {
		parts := strings.Split(lastDoc.NomorSurat, "/")
		if len(parts) > 1 {
			if num, err := strconv.Atoi(parts[1]); err == nil {
				lastNumFromDB = num
			}
		}
	}

	configRowTahun, _ := s.configRepo.GetForUpdate(tx, "nomor_surat_tahun_terakhir")
	configRowNomor, _ := s.configRepo.GetForUpdate(tx, "nomor_surat_terakhir")

	lastNumFromConfig := 0
	lastYearFromConfig := 0

	if configRowNomor != nil {
		lastNumFromConfig, _ = strconv.Atoi(configRowNomor.Value)
	}
	if configRowTahun != nil {
		lastYearFromConfig, _ = strconv.Atoi(configRowTahun.Value)
	}

	if year != lastYearFromConfig {
		lastNumFromConfig = 0
	}
	
	runningNumber := 0
	if lastNumFromDB > lastNumFromConfig {
		runningNumber = lastNumFromDB + 1
	} else {
		runningNumber = lastNumFromConfig + 1
	}
	
	newNomorStr := strconv.Itoa(runningNumber)
	newTahunStr := strconv.Itoa(year)
	
	configsToSet := map[string]string{
		"nomor_surat_terakhir":  newNomorStr,
		"nomor_surat_tahun_terakhir": newTahunStr,
	}

	if err := s.configRepo.SetMultiple(tx, configsToSet); err != nil {
		return "", "", fmt.Errorf("gagal update config nomor surat: %w", err)
	}
	
	docNumber := fmt.Sprintf(format, runningNumber, monthRoman, year)
	return docNumber, newNomorStr, nil
}

func (s *lostDocumentService) CreateLostDocument(residentData models.Resident, items []models.LostItem, operatorID uint, lokasiHilang string, petugasPelaporID uint, pejabatPersetujuID uint) (*models.LostDocument, error) {
	var createdDocID uint
	var finalDocNumber string
	
	s.docNumMutex.Lock()
	defer s.docNumMutex.Unlock()

	err := s.db.Transaction(func(tx *gorm.DB) error {
		var existingResident models.Resident
		err := tx.Where("nama_lengkap = ? AND tanggal_lahir = ?", residentData.NamaLengkap, residentData.TanggalLahir).First(&existingResident).Error
		
		if err == gorm.ErrRecordNotFound {
			residentData.NIK = s.generateTempNIK()
			newResident, createErr := s.residentRepo.Create(tx, &residentData)
			if createErr != nil {
				return createErr
			}
			existingResident = *newResident
		} else if err != nil {
			return err
		}

		docNumber, newNomor, err := s.generateDocumentNumber(tx)
		if err != nil {
			return err
		}
		finalDocNumber = docNumber
		
		loc, err := s.configService.GetLocation()
		if err != nil { loc = time.UTC }
		now := time.Now().In(loc)
		
		newDoc := &models.LostDocument{
			NomorSurat:         docNumber,
			TanggalLaporan:     now,
			Status:             "DITERBITKAN",
			LokasiHilang:       lokasiHilang,
			ResidentID:         existingResident.ID,
			PetugasPelaporID:   petugasPelaporID,
			PejabatPersetujuID: &pejabatPersetujuID,
			OperatorID:         operatorID,
			TanggalPersetujuan: &now,
			LostItems:          items,
		}
		
		created, err := s.docRepo.Create(tx, newDoc)
		if err != nil {
			return err
		}
		createdDocID = created.ID
		
		log.Printf("INFO: Nomor surat baru: %s (Nomor urut: %s)", finalDocNumber, newNomor)
		return nil
	})
	
	if err != nil {
		return nil, err
	}
	
	s.auditService.LogActivity(operatorID, models.AuditCreateDocument, fmt.Sprintf("Membuat surat keterangan hilang baru dengan nomor: %s", finalDocNumber))
	
	finalDoc, err := s.docRepo.FindByID(createdDocID)
	if err != nil {
		return nil, err
	}
	return finalDoc, nil
}

func (s *lostDocumentService) UpdateLostDocument(docID uint, residentData models.Resident, items []models.LostItem, lokasiHilang string, petugasPelaporID uint, pejabatPersetujuID uint, loggedInUserID uint) (*models.LostDocument, error) {
	var updatedDoc *models.LostDocument
	err := s.db.Transaction(func(tx *gorm.DB) error {
		existingDoc, err := s.docRepo.FindByID(docID)
		if err != nil { return err }

		loggedInUser, err := s.userRepo.FindByID(loggedInUserID)
		if err != nil { return errors.New("pengguna tidak valid") }

		if loggedInUser.Peran != models.RoleSuperAdmin && existingDoc.OperatorID != loggedInUserID {
			return ErrAccessDenied
		}
		
		if !strings.HasPrefix(existingDoc.Resident.NIK, "TEMP") {
			residentData.NIK = existingDoc.Resident.NIK
		} else {
			if residentData.NIK == "" {
				residentData.NIK = s.generateTempNIK()
			}
		}
		
		existingDoc.Resident.NamaLengkap = residentData.NamaLengkap
		existingDoc.Resident.TempatLahir = residentData.TempatLahir
		existingDoc.Resident.TanggalLahir = residentData.TanggalLahir
		existingDoc.Resident.JenisKelamin = residentData.JenisKelamin
		existingDoc.Resident.Agama = residentData.Agama
		existingDoc.Resident.Pekerjaan = residentData.Pekerjaan
		existingDoc.Resident.Alamat = residentData.Alamat
		existingDoc.Resident.NIK = residentData.NIK
		
		existingDoc.LokasiHilang = lokasiHilang
		existingDoc.PetugasPelaporID = petugasPelaporID
		existingDoc.PejabatPersetujuID = &pejabatPersetujuID
		existingDoc.LastUpdatedByID = &loggedInUserID
		
		if err := tx.Where("lost_document_id = ?", docID).Delete(&models.LostItem{}).Error; err != nil {
			return err
		}
		existingDoc.LostItems = items
		updatedDoc, err = s.docRepo.Update(tx, existingDoc)
		if err != nil { return err }
		return nil
	})
	if err != nil { return nil, err }

	s.auditService.LogActivity(loggedInUserID, models.AuditUpdateDocument, fmt.Sprintf("Memperbarui dokumen dengan Nomor Surat: %s", updatedDoc.NomorSurat))
	return updatedDoc, nil
}

func (s *lostDocumentService) DeleteLostDocument(id uint, loggedInUserID uint) error {
	var docToDelete models.LostDocument
	if err := s.db.First(&docToDelete, id).Error; err != nil {
		return errors.New("dokumen tidak ditemukan")
	}
	originalNomorSurat := docToDelete.NomorSurat
	err := s.db.Transaction(func(tx *gorm.DB) error {
		loggedInUser, err := s.userRepo.FindByID(loggedInUserID)
		if err != nil { return errors.New("pengguna tidak valid") }

		if loggedInUser.Peran != models.RoleSuperAdmin && docToDelete.OperatorID != loggedInUserID {
			return ErrAccessDenied
		}
		modifiedNomorSurat := fmt.Sprintf("DELETED_%d_%s", time.Now().Unix(), docToDelete.NomorSurat)
		if err := tx.Model(&models.LostDocument{}).Where("id = ?", id).Update("nomor_surat", modifiedNomorSurat).Error; err != nil {
			return err
		}
		if err := tx.Delete(&models.LostDocument{}, id).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil { return err }
	s.auditService.LogActivity(loggedInUserID, models.AuditDeleteDocument, fmt.Sprintf("Menghapus dokumen dengan Nomor Surat: %s", originalNomorSurat))
	return nil
}

func (s *lostDocumentService) processDocsStatus(docs []models.LostDocument) ([]models.LostDocument, error) {
	appConfig, err := s.configService.GetConfig()
	if err != nil { return nil, err }

	if appConfig == nil || appConfig.ArchiveDurationDays == 0 {
		return docs, nil
	}

	archiveDuration := time.Duration(appConfig.ArchiveDurationDays) * 24 * time.Hour

	for i := range docs {
		if docs[i].Status == "DITERBITKAN" && time.Now().After(docs[i].TanggalLaporan.Add(archiveDuration)) {
			docs[i].Status = "DIARSIPKAN"
		}
	}
	return docs, nil
}

func (s *lostDocumentService) SearchGlobal(query string) ([]models.LostDocument, error) {
	docs, err := s.docRepo.SearchGlobal(query)
	if err != nil { return nil, err }
	return s.processDocsStatus(docs)
}

func (s *lostDocumentService) FindAll(query string, statusFilter string) ([]models.LostDocument, error) {
	appConfig, err := s.configService.GetConfig()
	if err != nil { return nil, err }
	
	archiveDays := 15
	if appConfig != nil && appConfig.ArchiveDurationDays > 0 {
		archiveDays = appConfig.ArchiveDurationDays
	}

	docs, err := s.docRepo.FindAll(query, statusFilter, archiveDays)
	if err != nil { return nil, err }
	return s.processDocsStatus(docs)
}

func intToRoman(num int) string {
	romanNumeralMap := map[int]string{1: "I", 2: "II", 3: "III", 4: "IV", 5: "V", 6: "VI", 7: "VII", 8: "VIII", 9: "IX", 10: "X", 11: "XI", 12: "XII"}
	if val, ok := romanNumeralMap[num]; ok { return val }
	return ""
}