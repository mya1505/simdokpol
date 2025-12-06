package services

import (
	"errors"
	"simdokpol/internal/repositories"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// VAR JWTSecretKey SUDAH DIHAPUS DARI SINI (Pindah ke vars.go)

type AuthService interface {
	Login(nrp string, password string) (string, error)
}

type authService struct {
	userRepo      repositories.UserRepository
	configService ConfigService 
}

func NewAuthService(userRepo repositories.UserRepository, configService ConfigService) AuthService {
	return &authService{
		userRepo:      userRepo,
		configService: configService,
	}
}

func (s *authService) Login(nrp string, password string) (string, error) {
	user, err := s.userRepo.FindByNRP(nrp)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", errors.New("NRP atau kata sandi salah")
		}
		return "", err
	}

	if user.DeletedAt.Valid {
		return "", errors.New("akun Anda tidak aktif. Silakan hubungi Super Admin")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.KataSandi), []byte(password))
	if err != nil {
		return "", errors.New("NRP atau kata sandi salah")
	}

	config, _ := s.configService.GetConfig()
	timeoutMinutes := 480 // Default 8 Jam
	if config != nil && config.SessionTimeout > 0 {
		timeoutMinutes = config.SessionTimeout
	}
	
	expirationTime := time.Now().Add(time.Duration(timeoutMinutes) * time.Minute)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": user.ID,
		"role":   user.Peran,
		"exp":    expirationTime.Unix(),
	})
	
	// Gunakan variabel global dari vars.go
	tokenString, err := token.SignedString(JWTSecretKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}