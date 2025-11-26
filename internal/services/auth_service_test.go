package services

import (
	"errors"
	"simdokpol/internal/mocks"
	"simdokpol/internal/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert" // <-- PERBAIKAN TYPO "github.comcom"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func TestAuthService_Login(t *testing.T) {
	password := "password123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	assert.NoError(t, err)

	mockUser := &models.User{
		ID:        1,
		NRP:       "12345",
		KataSandi: string(hashedPassword),
		Peran:     models.RoleOperator,
		DeletedAt: gorm.DeletedAt{},
	}

	mockInactiveUser := &models.User{
		ID:        2,
		NRP:       "54321",
		KataSandi: string(hashedPassword),
		Peran:     models.RoleOperator,
		DeletedAt: gorm.DeletedAt{Time: time.Now(), Valid: true},
	}

	JWTSecretKey = []byte("test-secret")

	testCases := []struct {
		name          string
		nrp           string
		password      string
		setupMock     func(mockRepo *mocks.UserRepository)
		expectToken   bool
		expectedError string
	}{
		{
			name:        "Login Berhasil",
			nrp:         "12345",
			password:    "password123",
			setupMock: func(mockRepo *mocks.UserRepository) {
				mockRepo.On("FindByNRP", "12345").Return(mockUser, nil)
			},
			expectToken:   true,
			expectedError: "",
		},
		{
			name:        "Gagal - Kata Sandi Salah",
			nrp:         "12345",
			password:    "password-salah",
			setupMock: func(mockRepo *mocks.UserRepository) {
				mockRepo.On("FindByNRP", "12345").Return(mockUser, nil)
			},
			expectToken:   false,
			expectedError: "NRP atau kata sandi salah",
		},
		{
			name:        "Gagal - Pengguna Tidak Ditemukan",
			nrp:         "00000",
			password:    "password123",
			setupMock: func(mockRepo *mocks.UserRepository) {
				mockRepo.On("FindByNRP", "00000").Return(nil, gorm.ErrRecordNotFound)
			},
			expectToken:   false,
			expectedError: "NRP atau kata sandi salah",
		},
		{
			name:        "Gagal - Akun Tidak Aktif",
			nrp:         "54321",
			password:    "password123",
			setupMock: func(mockRepo *mocks.UserRepository) {
				mockRepo.On("FindByNRP", "54321").Return(mockInactiveUser, nil)
			},
			expectToken:   false,
			// --- PERBAIKAN LINTER MISMATCH DI SINI ---
			expectedError: "akun Anda tidak aktif. Silakan hubungi Super Admin",
		},
		{
			name:        "Gagal - Error Database Lainnya",
			nrp:         "12345",
			password:    "password123",
			setupMock: func(mockRepo *mocks.UserRepository) {
				mockRepo.On("FindByNRP", "12345").Return(nil, errors.New("koneksi database error"))
			},
			expectToken:   false,
			expectedError: "koneksi database error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockUserRepo := new(mocks.UserRepository)
			tc.setupMock(mockUserRepo)
			authService := NewAuthService(mockUserRepo)
			token, err := authService.Login(tc.nrp, tc.password)

			if tc.expectToken {
				assert.NoError(t, err, "Seharusnya tidak ada error")
				assert.NotEmpty(t, token, "Token seharusnya tidak kosong")
			} else {
				assert.Error(t, err, "Seharusnya ada error")
				assert.Empty(t, token, "Token seharusnya kosong")
				assert.Equal(t, tc.expectedError, err.Error(), "Pesan error tidak sesuai")
			}
			
			mockUserRepo.AssertExpectations(t)
		})
	}
}