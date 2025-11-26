// internal/mocks/user_service_mock.go
package mocks

import (
	"simdokpol/internal/models"

	"github.com/stretchr/testify/mock"
)

// UserService adalah mock lengkap untuk interface service pengguna.
type UserService struct {
	mock.Mock
}

// Create mock
func (m *UserService) Create(user *models.User, actorID uint) error {
	args := m.Called(user, actorID)
	return args.Error(0)
}

// UpdateProfile mock
func (m *UserService) UpdateProfile(userID uint, user *models.User) (*models.User, error) {
	args := m.Called(userID, user)
	if u := args.Get(0); u != nil {
		if usr, ok := u.(*models.User); ok {
			return usr, args.Error(1)
		}
	}
	return nil, args.Error(1)
}

// --- METHOD YANG HILANG DITAMBAHKAN DI BAWAH ---

func (m *UserService) FindAll(statusFilter string) ([]models.User, error) {
	args := m.Called(statusFilter)
	if u := args.Get(0); u != nil {
		if users, ok := u.([]models.User); ok {
			return users, args.Error(1)
		}
	}
	return nil, args.Error(1)
}

func (m *UserService) FindByID(id uint) (*models.User, error) {
	args := m.Called(id)
	if u := args.Get(0); u != nil {
		if user, ok := u.(*models.User); ok {
			return user, args.Error(1)
		}
	}
	return nil, args.Error(1)
}

func (m *UserService) FindOperators() ([]models.User, error) {
	args := m.Called()
	if u := args.Get(0); u != nil {
		if users, ok := u.([]models.User); ok {
			return users, args.Error(1)
		}
	}
	return nil, args.Error(1)
}

func (m *UserService) Update(user *models.User, newPassword string, actorID uint) error {
	args := m.Called(user, newPassword, actorID)
	return args.Error(0)
}

func (m *UserService) Deactivate(id uint, actorID uint) error {
	args := m.Called(id, actorID)
	return args.Error(0)
}

func (m *UserService) Activate(id uint, actorID uint) error {
	args := m.Called(id, actorID)
	return args.Error(0)
}

func (m *UserService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	args := m.Called(userID, oldPassword, newPassword)
	return args.Error(0)
}