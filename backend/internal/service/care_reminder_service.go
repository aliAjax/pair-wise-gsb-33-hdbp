package service

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/gbplantwiki/gbplantwiki/internal/constants"
	"github.com/gbplantwiki/gbplantwiki/internal/model"
	"github.com/gbplantwiki/gbplantwiki/internal/repository"
	"github.com/gbplantwiki/gbplantwiki/internal/util"
)

// CareReminderService implements care reminder state machine logic.
// The state machine (pending -> done / overdue) is intentionally mirrored in
// frontend button visibility, log templates, error codes and formatters.
// Reminders may be attached to a plant profile (garden_id); such reminders
// follow the profile across relocations and survive its removal.
type CareReminderService struct {
	repo       *repository.CareReminderRepository
	gardenRepo *repository.UserGardenRepository
	logger     *slog.Logger
}

// NewCareReminderService creates a CareReminderService.
func NewCareReminderService(repo *repository.CareReminderRepository, gardenRepo *repository.UserGardenRepository, logger *slog.Logger) *CareReminderService {
	return &CareReminderService{repo: repo, gardenRepo: gardenRepo, logger: logger}
}

// Create adds a reminder for the current user. When a garden id is given the
// reminder is anchored to that plant profile and inherits its species.
func (s *CareReminderService) Create(userID uint, m *model.CareReminder) (*model.CareReminder, error) {
	m.UserID = userID
	if m.Status == "" {
		m.Status = model.ReminderPending
	}
	if m.RemindDate.IsZero() {
		return nil, util.NewAppError(422, constants.CodeValidationError,
			fmt.Sprintf("CareReminder[task_title=%s] create failed: remind_date required", m.TaskTitle))
	}
	if m.GardenID > 0 {
		g, err := s.gardenRepo.FindByID(m.GardenID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return nil, util.NewAppError(404, constants.CodeNotFound, fmt.Sprintf("UserGarden[id=%d] not found", m.GardenID))
			}
			return nil, fmt.Errorf("care reminder garden find: %w", err)
		}
		if g.UserID != userID {
			return nil, util.NewAppError(403, constants.CodeForbidden,
				fmt.Sprintf("CareReminder[task_title=%s] create failed: garden_id=%d not owner", m.TaskTitle, m.GardenID))
		}
		m.PlantSpeciesID = g.PlantSpeciesID
	}
	if err := s.repo.Create(m); err != nil {
		s.logger.Error(fmt.Sprintf(constants.LogReminderCreateFailed, m.TaskTitle), "error", err)
		return nil, fmt.Errorf("care reminder create: %w", err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogReminderCreateSuccess, m.TaskTitle), "id", m.ID, "garden_id", m.GardenID)
	return m, nil
}

// ListByUser lists reminders with optional status and garden location filters.
func (s *CareReminderService) ListByUser(userID uint, status, location string) ([]model.CareReminderWithPlant, error) {
	if location != "" && !constants.IsValidGardenLocation(location) {
		return nil, util.NewAppError(422, constants.CodeValidationError,
			fmt.Sprintf("CareReminder list failed: unknown location %s", location))
	}
	if _, err := s.repo.MarkOverdue(userID); err != nil {
		s.logger.Warn("care reminder overdue mark failed", "error", err)
	}
	items, err := s.repo.ListByUser(userID, status, location)
	if err != nil {
		return nil, fmt.Errorf("care reminder list: %w", err)
	}
	return items, nil
}

// ListByMonth lists reminders within a calendar month.
func (s *CareReminderService) ListByMonth(userID uint, year, month int) ([]model.CareReminder, error) {
	items, err := s.repo.ListByMonth(userID, year, month)
	if err != nil {
		return nil, fmt.Errorf("care reminder month list: %w", err)
	}
	return items, nil
}

// UpdateStatus transitions a reminder to a new status.
func (s *CareReminderService) UpdateStatus(userID, id uint, status string) (*model.CareReminder, error) {
	m, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(404, constants.CodeNotFound, fmt.Sprintf("CareReminder[id=%d] not found", id))
		}
		return nil, fmt.Errorf("care reminder status find: %w", err)
	}
	if m.UserID != userID {
		return nil, util.NewAppError(403, constants.CodeForbidden,
			fmt.Sprintf("CareReminder[id=%d] status change failed: user_id=%d not owner", id, userID))
	}
	switch status {
	case model.ReminderDone:
		m.Status = model.ReminderDone
	case model.ReminderPending:
		if m.RemindDate.Before(time.Now()) {
			m.Status = model.ReminderOverdue
		} else {
			m.Status = model.ReminderPending
		}
	default:
		return nil, util.NewAppError(422, constants.CodeValidationError,
			fmt.Sprintf("CareReminder[id=%d] status=%s invalid transition", id, status))
	}
	if err := s.repo.Update(m); err != nil {
		return nil, fmt.Errorf("care reminder status update: %w", err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogReminderStatusChanged, id, m.Status), "id", id)
	return m, nil
}

// Delete removes a reminder owned by the user.
func (s *CareReminderService) Delete(userID, id uint) error {
	m, err := s.repo.FindByID(id)
	if err != nil {
		return fmt.Errorf("care reminder delete find: %w", err)
	}
	if m.UserID != userID {
		return util.NewAppError(403, constants.CodeForbidden, fmt.Sprintf("CareReminder[id=%d] delete failed: not owner", id))
	}
	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("care reminder delete: %w", err)
	}
	return nil
}
