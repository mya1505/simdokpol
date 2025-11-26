package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	// "simdokpol/internal/middleware" // <-- PERBAIKAN: HAPUS BARIS INI
	"simdokpol/internal/mocks"
	"simdokpol/internal/models"
	"simdokpol/internal/services"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// setupDocTestRouter membuat instance Gin untuk LostDocumentController
func setupDocTestRouter(mockDocService *mocks.LostDocumentService, authInjector gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)

	if mockDocService == nil {
		mockDocService = new(mocks.LostDocumentService)
	}

	docController := NewLostDocumentController(mockDocService)

	router := gin.New()
	if authInjector != nil {
		router.Use(authInjector)
	}

	// Rute dokumen standar (memerlukan auth biasa)
	docRoutes := router.Group("/api")
	{
		docRoutes.POST("/documents", docController.Create)
		docRoutes.GET("/documents", docController.FindAll)
		docRoutes.GET("/documents/:id", docController.FindByID)
		docRoutes.PUT("/documents/:id", docController.Update)
		docRoutes.DELETE("/documents/:id", docController.Delete)
		docRoutes.GET("/search", docController.SearchGlobal)
	}

	return router
}

// --- Helper Data ---
var (
	adminUser = &models.User{ID: 1, NamaLengkap: "Admin", Peran: models.RoleSuperAdmin}
	opOwner   = &models.User{ID: 2, NamaLengkap: "Operator Pemilik", Peran: models.RoleOperator}
	opOther   = &models.User{ID: 3, NamaLengkap: "Operator Lain", Peran: models.RoleOperator}

	mockDoc = &models.LostDocument{
		ID:         101,
		NomorSurat: "SKH/101/XI/TUK.7.2.1/2025",
		OperatorID: opOwner.ID, // Dimiliki oleh user ID 2
		Resident:   models.Resident{NamaLengkap: "BUDI SANTOSO"},
	}

	validDocRequest = DocumentRequest{
		NamaLengkap:        "BUDI SANTOSO",
		TempatLahir:        "JAKARTA",
		TanggalLahir:       "1990-01-15",
		JenisKelamin:       "Laki-laki",
		Agama:              "Islam",
		Pekerjaan:          "Karyawan Swasta",
		Alamat:             "JL. MERDEKA NO. 10, JAKARTA",
		LokasiHilang:       "Sekitar Pasar Senen",
		PetugasPelaporID:   2,
		PejabatPersetujuID: 1,
		Items: []struct {
			NamaBarang string `json:"nama_barang" binding:"required" example:"KTP"`
			Deskripsi  string `json:"deskripsi" example:"NIK: 3171234567890001"`
		}{
			{NamaBarang: "KTP", Deskripsi: "NIK: 3171234567890001"},
		},
	}
	
	expectedDocJSON = `{"id":101, "nomor_surat":"SKH/101/XI/TUK.7.2.1/2025", "tanggal_laporan":"0001-01-01T00:00:00Z", "status":"", "lokasi_hilang":"", "resident_id":0, "resident":{"id":0, "nik":"", "nama_lengkap":"BUDI SANTOSO", "tempat_lahir":"", "tanggal_lahir":"0001-01-01T00:00:00Z", "jenis_kelamin":"", "agama":"", "pekerjaan":"", "alamat":"", "created_at":"0001-01-01T00:00:00Z", "updated_at":"0001-01-01T00:00:00Z"}, "lost_items":null, "petugas_pelapor_id":0, "petugas_pelapor":{"id":0, "nama_lengkap":"", "nrp":"", "pangkat":"", "peran":"", "jabatan":"", "regu":"", "created_at":"0001-01-01T00:00:00Z", "updated_at":"0001-01-01T00:00:00Z"}, "pejabat_persetuju_id":null, "pejabat_persetuju":{"id":0, "nama_lengkap":"", "nrp":"", "pangkat":"", "peran":"", "jabatan":"", "regu":"", "created_at":"0001-01-01T00:00:00Z", "updated_at":"0001-01-01T00:00:00Z"}, "operator_id":2, "operator":{"id":0, "nama_lengkap":"", "nrp":"", "pangkat":"", "peran":"", "jabatan":"", "regu":"", "created_at":"0001-01-01T00:00:00Z", "updated_at":"0001-01-01T00:00:00Z"}, "last_updated_by_id":null, "last_updated_by":{"id":0, "nama_lengkap":"", "nrp":"", "pangkat":"", "peran":"", "jabatan":"", "regu":"", "created_at":"0001-01-01T00:00:00Z", "updated_at":"0001-01-01T00:00:00Z"}, "tanggal_persetujuan":null, "created_at":"0001-01-01T00:00:00Z", "updated_at":"0001-01-01T00:00:00Z"}`

)

// TestLostDocumentController_Create menguji endpoint POST /api/documents
func TestLostDocumentController_Create(t *testing.T) {
	testCases := []struct {
		name               string
		userInContext      *models.User
		requestBody        interface{}
		mockSetup          func(*mocks.LostDocumentService)
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name:          "Success - Create Document",
			userInContext: adminUser,
			requestBody:   validDocRequest,
			mockSetup: func(mockSvc *mocks.LostDocumentService) {
				mockSvc.On("CreateLostDocument",
					mock.AnythingOfType("models.Resident"),
					mock.AnythingOfType("[]models.LostItem"),
					adminUser.ID,
					validDocRequest.LokasiHilang,
					validDocRequest.PetugasPelaporID,
					validDocRequest.PejabatPersetujuID,
				).Return(mockDoc, nil).Once()
			},
			expectedStatusCode: http.StatusCreated,
			expectedBody:       expectedDocJSON,
		},
		{
			name:          "Failure - Invalid Request Body (Missing Items)",
			userInContext: adminUser,
			requestBody: gin.H{
				"nama_lengkap":         "BUDI SANTOSO",
				"tempat_lahir":         "JAKARTA",
				"tanggal_lahir":        "1990-01-15",
				"jenis_kelamin":        "Laki-laki",
				"agama":                "Islam",
				"pekerjaan":            "Karyawan Swasta",
				"alamat":               "JL. MERDEKA NO. 10, JAKARTA",
				"lokasi_hilang":        "Sekitar Pasar Senen",
				"petugas_pelapor_id":   2,
				"pejabat_persetuju_id": 1,
				"items":                []string{},
			},
			mockSetup:          func(mockSvc *mocks.LostDocumentService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       `{"error":"Input tidak valid: Key: 'DocumentRequest.Items' Error:Field validation for 'Items' failed on the 'min' tag"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDocService := new(mocks.LostDocumentService)
			authInjector := func(c *gin.Context) {
				c.Set("currentUser", tc.userInContext)
				c.Set("userID", tc.userInContext.ID)
				c.Next()
			}
			router := setupDocTestRouter(mockDocService, authInjector)
			tc.mockSetup(mockDocService)

			jsonBody, err := json.Marshal(tc.requestBody)
			assert.NoError(t, err)

			req, _ := http.NewRequest(http.MethodPost, "/api/documents", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "application/json")
			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, req)

			assert.Equal(t, tc.expectedStatusCode, recorder.Code)
			assert.JSONEq(t, tc.expectedBody, recorder.Body.String())
			mockDocService.AssertExpectations(t)
		})
	}
}

// TestLostDocumentController_FindByID_Authorization menguji otorisasi endpoint GET /api/documents/:id
func TestLostDocumentController_FindByID_Authorization(t *testing.T) {
	testCases := []struct {
		name               string
		userInContext      *models.User
		docID              string
		mockSetup          func(*mocks.LostDocumentService)
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name:          "Success - Super Admin can access any document",
			userInContext: adminUser,
			docID:         "101",
			mockSetup: func(mockSvc *mocks.LostDocumentService) {
				mockSvc.On("FindByID", uint(101), adminUser.ID).Return(mockDoc, nil).Once()
			},
			expectedStatusCode: http.StatusOK,
			expectedBody:       expectedDocJSON,
		},
		{
			name:          "Success - Operator can access own document",
			userInContext: opOwner,
			docID:         "101",
			mockSetup: func(mockSvc *mocks.LostDocumentService) {
				mockSvc.On("FindByID", uint(101), opOwner.ID).Return(mockDoc, nil).Once()
			},
			expectedStatusCode: http.StatusOK,
			expectedBody:       expectedDocJSON,
		},
		{
			name:          "Failure - Operator cannot access other's document",
			userInContext: opOther,
			docID:         "101",
			mockSetup: func(mockSvc *mocks.LostDocumentService) {
				mockSvc.On("FindByID", uint(101), opOther.ID).Return(nil, services.ErrAccessDenied).Once()
			},
			expectedStatusCode: http.StatusForbidden,
			expectedBody:       `{"error":"Akses ditolak: Anda tidak memiliki izin untuk melihat dokumen ini."}`,
		},
		{
			name:          "Failure - Document Not Found",
			userInContext: adminUser,
			docID:         "999",
			mockSetup: func(mockSvc *mocks.LostDocumentService) {
				mockSvc.On("FindByID", uint(999), adminUser.ID).Return(nil, errors.New("data tidak ditemukan")).Once()
			},
			expectedStatusCode: http.StatusNotFound,
			expectedBody:       `{"error":"Dokumen tidak ditemukan"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDocService := new(mocks.LostDocumentService)
			authInjector := func(c *gin.Context) {
				c.Set("currentUser", tc.userInContext)
				c.Set("userID", tc.userInContext.ID)
				c.Next()
			}
			router := setupDocTestRouter(mockDocService, authInjector)
			tc.mockSetup(mockDocService)

			req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/documents/%s", tc.docID), nil)
			req.Header.Set("Accept", "application/json")
			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, req)

			assert.Equal(t, tc.expectedStatusCode, recorder.Code)
			assert.JSONEq(t, tc.expectedBody, recorder.Body.String())
			mockDocService.AssertExpectations(t)
		})
	}
}

// TestLostDocumentController_Update_Authorization
func TestLostDocumentController_Update_Authorization(t *testing.T) {
	testCases := []struct {
		name               string
		userInContext      *models.User
		docID              string
		requestBody        interface{}
		mockSetup          func(*mocks.LostDocumentService)
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name:          "Success - Super Admin can update any document",
			userInContext: adminUser,
			docID:         "101",
			requestBody:   validDocRequest,
			mockSetup: func(mockSvc *mocks.LostDocumentService) {
				mockSvc.On("UpdateLostDocument", uint(101),
					mock.AnythingOfType("models.Resident"),
					mock.AnythingOfType("[]models.LostItem"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("uint"),
					mock.AnythingOfType("uint"),
					adminUser.ID,
				).Return(mockDoc, nil).Once()
			},
			expectedStatusCode: http.StatusOK,
			expectedBody:       expectedDocJSON,
		},
		{
			name:          "Success - Operator can update own document",
			userInContext: opOwner,
			docID:         "101",
			requestBody:   validDocRequest,
			mockSetup: func(mockSvc *mocks.LostDocumentService) {
				mockSvc.On("UpdateLostDocument", uint(101),
					mock.AnythingOfType("models.Resident"),
					mock.AnythingOfType("[]models.LostItem"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("uint"),
					mock.AnythingOfType("uint"),
					opOwner.ID,
				).Return(mockDoc, nil).Once()
			},
			expectedStatusCode: http.StatusOK,
			expectedBody:       expectedDocJSON,
		},
		{
			name:          "Failure - Operator cannot update other's document",
			userInContext: opOther,
			docID:         "101",
			requestBody:   validDocRequest,
			mockSetup: func(mockSvc *mocks.LostDocumentService) {
				mockSvc.On("UpdateLostDocument", uint(101),
					mock.AnythingOfType("models.Resident"),
					mock.AnythingOfType("[]models.LostItem"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("uint"),
					mock.AnythingOfType("uint"),
					opOther.ID,
				).Return(nil, services.ErrAccessDenied).Once()
			},
			expectedStatusCode: http.StatusForbidden,
			expectedBody:       `{"error":"akses ditolak"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDocService := new(mocks.LostDocumentService)
			authInjector := func(c *gin.Context) {
				c.Set("currentUser", tc.userInContext)
				c.Set("userID", tc.userInContext.ID)
				c.Next()
			}
			router := setupDocTestRouter(mockDocService, authInjector)
			tc.mockSetup(mockDocService)

			jsonBody, err := json.Marshal(tc.requestBody)
			assert.NoError(t, err)

			req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/documents/%s", tc.docID), bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "application/json")
			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, req)

			assert.Equal(t, tc.expectedStatusCode, recorder.Code)
			if tc.expectedStatusCode != http.StatusOK {
				assert.JSONEq(t, tc.expectedBody, recorder.Body.String())
			}
			mockDocService.AssertExpectations(t)
		})
	}
}

// TestLostDocumentController_Delete_Authorization
func TestLostDocumentController_Delete_Authorization(t *testing.T) {
	testCases := []struct {
		name               string
		userInContext      *models.User
		docID              string
		mockSetup          func(*mocks.LostDocumentService)
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name:          "Success - Super Admin can delete any document",
			userInContext: adminUser,
			docID:         "101",
			mockSetup: func(mockSvc *mocks.LostDocumentService) {
				mockSvc.On("DeleteLostDocument", uint(101), adminUser.ID).Return(nil).Once()
			},
			expectedStatusCode: http.StatusOK,
			expectedBody:       `{"message":"Dokumen berhasil dihapus"}`,
		},
		{
			name:          "Success - Operator can delete own document",
			userInContext: opOwner,
			docID:         "101",
			mockSetup: func(mockSvc *mocks.LostDocumentService) {
				mockSvc.On("DeleteLostDocument", uint(101), opOwner.ID).Return(nil).Once()
			},
			expectedStatusCode: http.StatusOK,
			expectedBody:       `{"message":"Dokumen berhasil dihapus"}`,
		},
		{
			name:          "Failure - Operator cannot delete other's document",
			userInContext: opOther,
			docID:         "101",
			mockSetup: func(mockSvc *mocks.LostDocumentService) {
				mockSvc.On("DeleteLostDocument", uint(101), opOther.ID).Return(services.ErrAccessDenied).Once()
			},
			expectedStatusCode: http.StatusForbidden,
			expectedBody:       `{"error":"akses ditolak"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDocService := new(mocks.LostDocumentService)
			authInjector := func(c *gin.Context) {
				c.Set("currentUser", tc.userInContext)
				c.Set("userID", tc.userInContext.ID)
				c.Next()
			}
			router := setupDocTestRouter(mockDocService, authInjector)
			tc.mockSetup(mockDocService)

			req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/documents/%s", tc.docID), nil)
			req.Header.Set("Accept", "application/json")
			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, req)

			assert.Equal(t, tc.expectedStatusCode, recorder.Code)
			assert.JSONEq(t, tc.expectedBody, recorder.Body.String())
			mockDocService.AssertExpectations(t)
		})
	}
}

// TestLostDocumentController_FindAll
func TestLostDocumentController_FindAll(t *testing.T) {
	mockDocs := []models.LostDocument{*mockDoc}
	
	testCases := []struct {
		name               string
		userInContext      *models.User
		query              string
		mockSetup          func(*mocks.LostDocumentService)
		expectedStatusCode int
	}{
		{
			name:          "Success - Get active docs",
			userInContext: adminUser,
			query:         "?status=active",
			mockSetup: func(mockSvc *mocks.LostDocumentService) {
				mockSvc.On("FindAll", "", "active").Return(mockDocs, nil).Once()
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:          "Success - Get archived docs with query",
			userInContext: adminUser,
			query:         "?status=archived&q=Budi",
			mockSetup: func(mockSvc *mocks.LostDocumentService) {
				mockSvc.On("FindAll", "Budi", "archived").Return(mockDocs, nil).Once()
			},
			expectedStatusCode: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDocService := new(mocks.LostDocumentService)
			authInjector := func(c *gin.Context) {
				c.Set("currentUser", tc.userInContext); c.Set("userID", tc.userInContext.ID); c.Next()
			}
			router := setupDocTestRouter(mockDocService, authInjector)
			tc.mockSetup(mockDocService)

			req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/documents%s", tc.query), nil)
			req.Header.Set("Accept", "application/json")
			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, req)

			assert.Equal(t, tc.expectedStatusCode, recorder.Code)
			if tc.expectedStatusCode == http.StatusOK {
				assert.Contains(t, recorder.Body.String(), "SKH/101/XI/TUK.7.2.1/2025")
			}
			mockDocService.AssertExpectations(t)
		})
	}
}

// TestLostDocumentController_SearchGlobal
func TestLostDocumentController_SearchGlobal(t *testing.T) {
	mockDocs := []models.LostDocument{*mockDoc}

	testCases := []struct {
		name               string
		userInContext      *models.User
		query              string
		mockSetup          func(*mocks.LostDocumentService)
		expectedStatusCode int
	}{
		{
			name:          "Success - Search with query",
			userInContext: adminUser,
			query:         "?q=Budi",
			mockSetup: func(mockSvc *mocks.LostDocumentService) {
				mockSvc.On("SearchGlobal", "Budi").Return(mockDocs, nil).Once()
			},
			expectedStatusCode: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDocService := new(mocks.LostDocumentService)
			authInjector := func(c *gin.Context) {
				c.Set("currentUser", tc.userInContext); c.Set("userID", tc.userInContext.ID); c.Next()
			}
			router := setupDocTestRouter(mockDocService, authInjector)
			tc.mockSetup(mockDocService)

			req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/search%s", tc.query), nil)
			req.Header.Set("Accept", "application/json")
			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, req)

			assert.Equal(t, tc.expectedStatusCode, recorder.Code)
			if tc.expectedStatusCode == http.StatusOK {
				assert.Contains(t, recorder.Body.String(), "SKH/101/XI/TUK.7.2.1/2025")
			}
			mockDocService.AssertExpectations(t)
		})
	}
}