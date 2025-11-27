package services

import (
	"errors"
	"fmt"
	"simdokpol/internal/config"
	"simdokpol/internal/dto" // <-- Pastikan import ini
	"simdokpol/internal/models"
	"simdokpol/internal/repositories"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Create(user *models.User, actorID uint) error
	// Update: Ganti FindAll jadi GetUsersPaged
	GetUsersPaged(req dto.DataTableRequest, statusFilter string) (*dto.DataTableResponse, error)
	FindByID(id uint) (*models.User, error)
	FindOperators() ([]models.User, error)
	Update(user *models.User, newPassword string, actorID uint) error
	Deactivate(id uint, actorID uint) error
	Activate(id uint, actorID uint) error
	ChangePassword(userID uint, oldPassword, newPassword string) error
	UpdateProfile(userID uint, dataToUpdate *models.User) (*models.User, error)
}

type userService struct {
	userRepo     repositories.UserRepository
	auditService AuditLogService
	cfg          *config.Config
}

func NewUserService(userRepo repositories.UserRepository, auditService AuditLogService, cfg *config.Config) UserService {
	return &userService{
		userRepo:     userRepo,
		auditService: auditService,
		cfg:          cfg,
	}
}

// --- IMPLEMENTASI BARU ---
func (s *userService) GetUsersPaged(req dto.DataTableRequest, statusFilter string) (*dto.DataTableResponse, error) {
	users, total, filtered, err := s.userRepo.FindAllPaged(req, statusFilter)
	if err != nil {
		return nil, err
	}

	return &dto.DataTableResponse{
		Draw:            req.Draw,
		RecordsTotal:    total,
		RecordsFiltered: filtered,
		Data:            users,
	}, nil
}
// -------------------------

// ... (SISA FUNGSI LAINNYA: Create, Update, FindByID dll TETAP SAMA SEPERTI SEBELUMNYA) ...
// ... Copy Paste saja fungsi lama di bawah sini agar tidak hilang ...

func (s *userService) UpdateProfile(userID uint, dataToUpdate *models.User) (*models.User, error) {
	currentUser, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("pengguna tidak ditemukan")
	}
	currentUser.NamaLengkap = dataToUpdate.NamaLengkap
	currentUser.NRP = dataToUpdate.NRP
	currentUser.Pangkat = dataToUpdate.Pangkat

	if err := s.userRepo.Update(currentUser); err != nil {
		return nil, err
	}
	s.auditService.LogActivity(userID, models.AuditUpdateUser, fmt.Sprintf("Pengguna '%s' memperbarui profil.", currentUser.NamaLengkap))
	return currentUser, nil
}

func (s *userService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil { return errors.New("pengguna tidak ditemukan") }

	err = bcrypt.CompareHashAndPassword([]byte(user.KataSandi), []byte(oldPassword))
	if err != nil { return errors.New("kata sandi lama salah") }

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), s.cfg.BcryptCost)
	if err != nil { return err }
	user.KataSandi = string(hashedPassword)

	if err := s.userRepo.Update(user); err != nil { return err }
	s.auditService.LogActivity(userID, models.AuditUpdateUser, "Pengguna mengubah kata sandi.")
	return nil
}

func (s *userService) Create(user *models.User, actorID uint) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.KataSandi), s.cfg.BcryptCost)
	if err != nil { return err }
	user.KataSandi = string(hashedPassword)
	if err := s.userRepo.Create(user); err != nil { return err }
	
	action := models.AuditCreateUser
	if actorID == 0 { actorID = user.ID; action = models.AuditSystemSetup }
	
	s.auditService.LogActivity(actorID, action, fmt.Sprintf("Pengguna baru '%s' dibuat.", user.NamaLengkap))
	return nil
}

func (s *userService) Update(user *models.User, newPassword string, actorID uint) error {
	oldUser, err := s.userRepo.FindByID(user.ID)
	if err != nil { return errors.New("pengguna tidak ditemukan") }

	if strings.TrimSpace(newPassword) != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), s.cfg.BcryptCost)
		if err != nil { return err }
		user.KataSandi = string(hashedPassword)
	} else {
		user.KataSandi = oldUser.KataSandi
	}

	if err := s.userRepo.Update(user); err != nil { return err }
	s.auditService.LogActivity(actorID, models.AuditUpdateUser, fmt.Sprintf("Data pengguna '%s' diperbarui.", user.NamaLengkap))
	return nil
}

func (s *userService) Deactivate(id uint, actorID uint) error {
	user, err := s.userRepo.FindByID(id)
	if err != nil { return err }
	if err := s.userRepo.Delete(id); err != nil { return err }
	s.auditService.LogActivity(actorID, models.AuditDeactivateUser, fmt.Sprintf("Pengguna '%s' dinonaktifkan.", user.NamaLengkap))
	return nil
}

func (s *userService) Activate(id uint, actorID uint) error {
	user, err := s.userRepo.FindByID(id)
	if err != nil { return err }
	if err := s.userRepo.Restore(id); err != nil { return err }
	s.auditService.LogActivity(actorID, models.AuditActivateUser, fmt.Sprintf("Pengguna '%s' diaktifkan kembali.", user.NamaLengkap))
	return nil
}

func (s *userService) FindByID(id uint) (*models.User, error) {
	return s.userRepo.FindByID(id)
}

func (s *userService) FindOperators() ([]models.User, error) {
	return s.userRepo.FindOperators()
}