package services

import (
	"errors"
	"fmt"
	"simdokpol/internal/models"
	"simdokpol/internal/repositories"
	"strings"
)

type JobPositionService interface {
	FindAll() ([]models.JobPosition, error)
	FindAllActive() ([]models.JobPosition, error)
	Create(position *models.JobPosition, actorID uint) error
	Update(position *models.JobPosition, actorID uint) error
	Delete(id uint, actorID uint) error
	Restore(id uint, actorID uint) error
}

type jobPositionService struct {
	repo         repositories.JobPositionRepository
	auditService AuditLogService
}

func NewJobPositionService(repo repositories.JobPositionRepository, auditService AuditLogService) JobPositionService {
	return &jobPositionService{
		repo:         repo,
		auditService: auditService,
	}
}

func (s *jobPositionService) FindAll() ([]models.JobPosition, error) {
	return s.repo.FindAll()
}

func (s *jobPositionService) FindAllActive() ([]models.JobPosition, error) {
	return s.repo.FindAllActive()
}

func (s *jobPositionService) Create(position *models.JobPosition, actorID uint) error {
	if position == nil || position.Nama == "" {
		return errors.New("jabatan tidak valid")
	}
	position.Nama = strings.ToUpper(strings.TrimSpace(position.Nama))
	if err := s.repo.Create(position); err != nil {
		return err
	}
	s.auditService.LogActivity(actorID, models.AuditSettingsUpdated, fmt.Sprintf("Menambah jabatan: %s", position.Nama))
	return nil
}

func (s *jobPositionService) Update(position *models.JobPosition, actorID uint) error {
	if position == nil || position.Nama == "" {
		return errors.New("jabatan tidak valid")
	}
	position.Nama = strings.ToUpper(strings.TrimSpace(position.Nama))
	if err := s.repo.Update(position); err != nil {
		return err
	}
	s.auditService.LogActivity(actorID, models.AuditSettingsUpdated, fmt.Sprintf("Memperbarui jabatan: %s", position.Nama))
	return nil
}

func (s *jobPositionService) Delete(id uint, actorID uint) error {
	if err := s.repo.Delete(id); err != nil {
		return err
	}
	s.auditService.LogActivity(actorID, models.AuditSettingsUpdated, fmt.Sprintf("Menonaktifkan jabatan ID: %d", id))
	return nil
}

func (s *jobPositionService) Restore(id uint, actorID uint) error {
	if err := s.repo.Restore(id); err != nil {
		return err
	}
	s.auditService.LogActivity(actorID, models.AuditSettingsUpdated, fmt.Sprintf("Mengaktifkan jabatan ID: %d", id))
	return nil
}
