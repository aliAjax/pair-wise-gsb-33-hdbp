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
// Reminders bound to a garden item follow the plant: their location is always
// derived from the garden item, never stored on the reminder itself.
type CareReminderService struct {
	repo         *repository.CareReminderRepository
	gardenRepo   *repository.UserGardenRepository
	locationRepo *repository.GardenLocationRepository
	logger       *slog.Logger
}

// NewCareReminderService creates a CareReminderService.
func NewCareReminderService(repo *repository.CareReminderRepository, gardenRepo *repository.UserGardenRepository, locationRepo *repository.GardenLocationRepository, logger *slog.Logger) *CareReminderService {
	return &CareReminderService{repo: repo, gardenRepo: gardenRepo, locationRepo: locationRepo, logger: logger}
}

// Create adds a reminder for the current user. When GardenID is set the
// reminder is bound to that plant profile and inherits its species.
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
		if g.Status != constants.GardenStatusActive {
			return nil, util.NewAppError(422, constants.CodeValidationError,
				fmt.Sprintf("CareReminder[task_title=%s] create failed: UserGarden[id=%d] status=%s not active", m.TaskTitle, g.ID, g.Status))
		}
		m.PlantSpeciesID = g.PlantSpeciesID
	}
	if err := s.repo.Create(m); err != nil {
		s.logger.Error(fmt.Sprintf(constants.LogReminderCreateFailed, m.TaskTitle), "error", err)
		return nil, fmt.Errorf("care reminder create: %w", err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogReminderCreateSuccess, m.TaskTitle), "id", m.ID)
	return m, nil
}

// ListByUser lists reminders with status and location filters, enriched with
// the linked plant profile (garden code, nickname, current location).
func (s *CareReminderService) ListByUser(userID uint, status string, locationID uint) ([]model.CareReminder, error) {
	if _, err := s.repo.MarkOverdue(userID); err != nil {
		s.logger.Warn("care reminder overdue mark failed", "error", err)
	}
	items, err := s.repo.ListByUser(userID, status, locationID)
	if err != nil {
		return nil, fmt.Errorf("care reminder list: %w", err)
	}
	if err := s.fillGardenInfo(userID, items); err != nil {
		return nil, err
	}
	return items, nil
}

// ListByMonth lists reminders within a calendar month.
func (s *CareReminderService) ListByMonth(userID uint, year, month int) ([]model.CareReminder, error) {
	items, err := s.repo.ListByMonth(userID, year, month)
	if err != nil {
		return nil, fmt.Errorf("care reminder month list: %w", err)
	}
	if err := s.fillGardenInfo(userID, items); err != nil {
		return nil, err
	}
	return items, nil
}

// fillGardenInfo populates the transient plant profile fields of reminders.
// Garden items are loaded regardless of their status so completed reminders
// of removed plants still show which plant they belonged to.
func (s *CareReminderService) fillGardenInfo(userID uint, items []model.CareReminder) error {
	idSet := make(map[uint]bool)
	for _, m := range items {
		if m.GardenID > 0 {
			idSet[m.GardenID] = true
		}
	}
	if len(idSet) == 0 {
		return nil
	}
	ids := make([]uint, 0, len(idSet))
	for id := range idSet {
		ids = append(ids, id)
	}
	gardens, err := s.gardenRepo.FindByIDs(userID, ids)
	if err != nil {
		return fmt.Errorf("care reminder garden fill: %w", err)
	}
	gardenMap := make(map[uint]model.UserGarden, len(gardens))
	for _, g := range gardens {
		gardenMap[g.ID] = g
	}
	locs, err := s.locationRepo.ListByUser(userID)
	if err != nil {
		return fmt.Errorf("care reminder location fill: %w", err)
	}
	locNames := make(map[uint]string, len(locs))
	for _, loc := range locs {
		locNames[loc.ID] = loc.Name
	}
	for i := range items {
		g, ok := gardenMap[items[i].GardenID]
		if !ok {
			continue
		}
		items[i].GardenCode = g.GardenCode
		items[i].PlantNickname = g.Nickname
		items[i].LocationID = g.LocationID
		items[i].LocationName = locNames[g.LocationID]
	}
	return nil
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
