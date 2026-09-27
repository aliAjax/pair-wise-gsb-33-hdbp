package service

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/gbplantwiki/gbplantwiki/internal/constants"
	"github.com/gbplantwiki/gbplantwiki/internal/dto"
	"github.com/gbplantwiki/gbplantwiki/internal/model"
	"github.com/gbplantwiki/gbplantwiki/internal/repository"
	"github.com/gbplantwiki/gbplantwiki/internal/util"
)

// UserGardenService implements plant-profile logic: unique garden numbers,
// location capacities, relocations (single and atomic batch) and soft removal.
type UserGardenService struct {
	repo     *repository.UserGardenRepository
	moveRepo *repository.GardenMoveRepository
	capRepo  *repository.GardenLocationCapRepository
	logger   *slog.Logger
}

// NewUserGardenService creates a UserGardenService.
func NewUserGardenService(repo *repository.UserGardenRepository, moveRepo *repository.GardenMoveRepository,
	capRepo *repository.GardenLocationCapRepository, logger *slog.Logger) *UserGardenService {
	return &UserGardenService{repo: repo, moveRepo: moveRepo, capRepo: capRepo, logger: logger}
}

// Add creates a plant profile: validates the location, checks its remaining
// capacity and assigns the next unique garden number.
func (s *UserGardenService) Add(userID uint, g *model.UserGarden) (*model.UserGarden, error) {
	g.UserID = userID
	if g.Location == "" {
		g.Location = constants.GardenLocationIndoor
	}
	if !constants.IsValidGardenLocation(g.Location) {
		return nil, util.NewAppError(422, constants.CodeValidationError,
			fmt.Sprintf("UserGarden[location=%s] add failed: unknown location", g.Location))
	}
	if g.OwnedSince.IsZero() {
		g.OwnedSince = time.Now()
	}
	stats, err := s.LocationStats(userID)
	if err != nil {
		return nil, fmt.Errorf("user garden add stats: %w", err)
	}
	for _, st := range stats {
		if st.Location == g.Location && st.Remaining < 1 {
			return nil, util.NewAppError(409, constants.CodeConflict,
				fmt.Sprintf("%s容量不足：容量 %d，已用 %d，放不下新植株", locationText(st.Location), st.Capacity, st.Used))
		}
	}
	if err := s.createWithGardenNo(g); err != nil {
		s.logger.Error(fmt.Sprintf(constants.LogGardenAddFailed, g.PlantSpeciesID, userID), "error", err)
		return nil, fmt.Errorf("user garden add: %w", err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogGardenAddSuccess, g.PlantSpeciesID, userID), "id", g.ID, "garden_no", g.GardenNo)
	return g, nil
}

// createWithGardenNo assigns the next per-user garden number and retries on
// the unique (user_id, garden_no) index in case of concurrent creates.
func (s *UserGardenService) createWithGardenNo(g *model.UserGarden) error {
	total, err := s.repo.CountAllByUser(g.UserID)
	if err != nil {
		return err
	}
	for i := int64(1); i <= 5; i++ {
		g.GardenNo = fmt.Sprintf("G-%04d", total+i)
		if err := s.repo.Create(g); err != nil {
			if errors.Is(err, repository.ErrDuplicate) {
				continue
			}
			return err
		}
		return nil
	}
	return fmt.Errorf("allocate garden number for user %d: %w", g.UserID, repository.ErrConflict)
}

// List returns a user's active plant profiles, optionally filtered by location.
func (s *UserGardenService) List(userID uint, location string) ([]model.UserGarden, error) {
	if location != "" && !constants.IsValidGardenLocation(location) {
		return nil, util.NewAppError(422, constants.CodeValidationError,
			fmt.Sprintf("UserGarden list failed: unknown location %s", location))
	}
	items, err := s.repo.ListByUser(userID, location)
	if err != nil {
		return nil, fmt.Errorf("user garden list: %w", err)
	}
	return items, nil
}

// LocationStats reports capacity, usage and remaining slots of every location.
func (s *UserGardenService) LocationStats(userID uint) ([]dto.GardenLocationStat, error) {
	caps, err := s.capRepo.MapByUser(userID)
	if err != nil {
		return nil, fmt.Errorf("garden location caps: %w", err)
	}
	used, err := s.repo.CountByLocation(userID)
	if err != nil {
		return nil, fmt.Errorf("garden location usage: %w", err)
	}
	stats := make([]dto.GardenLocationStat, 0, len(constants.ValidGardenLocations()))
	for _, loc := range constants.ValidGardenLocations() {
		capacity, ok := caps[loc]
		if !ok {
			capacity = constants.DefaultGardenLocationCapacity(loc)
		}
		stats = append(stats, dto.GardenLocationStat{
			Location:  loc,
			Capacity:  capacity,
			Used:      used[loc],
			Remaining: capacity - used[loc],
		})
	}
	return stats, nil
}

// UpdateCapacity overrides the capacity of one location. The new capacity may
// not drop below the number of plants currently placed there.
func (s *UserGardenService) UpdateCapacity(userID uint, location string, capacity int) (*dto.GardenLocationStat, error) {
	if !constants.IsValidGardenLocation(location) {
		return nil, util.NewAppError(422, constants.CodeValidationError,
			fmt.Sprintf("GardenLocation[location=%s] capacity failed: unknown location", location))
	}
	used, err := s.repo.CountByLocation(userID)
	if err != nil {
		return nil, fmt.Errorf("garden capacity usage: %w", err)
	}
	if capacity < used[location] {
		return nil, util.NewAppError(409, constants.CodeConflict,
			fmt.Sprintf("%s容量不能小于当前已放置的 %d 株", locationText(location), used[location]))
	}
	if err := s.capRepo.Upsert(userID, location, capacity); err != nil {
		return nil, fmt.Errorf("garden capacity upsert: %w", err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogGardenCapacityUpdated, userID, location, capacity))
	return &dto.GardenLocationStat{Location: location, Capacity: capacity, Used: used[location], Remaining: capacity - used[location]}, nil
}

// Move relocates one plant profile after checking the target's remaining
// capacity, and records the relocation.
func (s *UserGardenService) Move(userID, id uint, toLocation string) (*model.UserGarden, error) {
	if !constants.IsValidGardenLocation(toLocation) {
		return nil, util.NewAppError(422, constants.CodeValidationError,
			fmt.Sprintf("UserGarden[id=%d] move failed: unknown location %s", id, toLocation))
	}
	item, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(404, constants.CodeNotFound, fmt.Sprintf("UserGarden[id=%d] not found", id))
		}
		return nil, fmt.Errorf("user garden move find: %w", err)
	}
	if item.UserID != userID {
		return nil, util.NewAppError(403, constants.CodeForbidden, fmt.Sprintf("UserGarden[id=%d] move failed: not owner", id))
	}
	if item.Location == toLocation {
		return item, nil
	}
	specs := []dto.GardenBatchMoveItem{{GardenID: id, Location: toLocation}}
	if err := s.applyMoves(userID, specs); err != nil {
		s.logger.Warn(fmt.Sprintf(constants.LogGardenMoveFailed, id, toLocation), "error", err)
		return nil, err
	}
	item.Location = toLocation
	return item, nil
}

// BatchMove relocates several plant profiles atomically: when any target
// location cannot hold its incoming plants the whole batch is rejected and
// the shortage per location is reported.
func (s *UserGardenService) BatchMove(userID uint, moves []dto.GardenBatchMoveItem) error {
	if err := s.applyMoves(userID, moves); err != nil {
		return err
	}
	return nil
}

// applyMoves validates and applies a set of relocations inside one transaction.
func (s *UserGardenService) applyMoves(userID uint, moves []dto.GardenBatchMoveItem) error {
	if len(moves) == 0 {
		return util.NewAppError(422, constants.CodeValidationError, "garden move failed: empty move list")
	}
	for _, m := range moves {
		if !constants.IsValidGardenLocation(m.Location) {
			return util.NewAppError(422, constants.CodeValidationError,
				fmt.Sprintf("garden move failed: unknown location %s", m.Location))
		}
	}
	// One plant can only land in one place: last entry wins on duplicates.
	seen := map[uint]int{}
	deduped := make([]dto.GardenBatchMoveItem, 0, len(moves))
	for _, m := range moves {
		if idx, ok := seen[m.GardenID]; ok {
			deduped[idx] = m
			continue
		}
		seen[m.GardenID] = len(deduped)
		deduped = append(deduped, m)
	}
	moves = deduped
	ids := make([]uint, 0, len(moves))
	for _, m := range moves {
		ids = append(ids, m.GardenID)
	}
	items, err := s.repo.FindActiveByIDs(userID, ids)
	if err != nil {
		return fmt.Errorf("garden move load: %w", err)
	}
	byID := make(map[uint]model.UserGarden, len(items))
	for _, it := range items {
		byID[it.ID] = it
	}
	for _, m := range moves {
		if _, ok := byID[m.GardenID]; !ok {
			return util.NewAppError(404, constants.CodeNotFound,
				fmt.Sprintf("UserGarden[id=%d] not found in current garden", m.GardenID))
		}
	}

	used, err := s.repo.CountByLocation(userID)
	if err != nil {
		return fmt.Errorf("garden move usage: %w", err)
	}
	caps, err := s.capRepo.MapByUser(userID)
	if err != nil {
		return fmt.Errorf("garden move caps: %w", err)
	}
	capOf := func(loc string) int {
		if c, ok := caps[loc]; ok {
			return c
		}
		return constants.DefaultGardenLocationCapacity(loc)
	}

	changes, shortages := planBatch(byID, moves, used, capOf)
	if len(shortages) > 0 {
		prefix := "搬位被拒绝："
		if len(moves) > 1 {
			prefix = "批量搬位被拒绝："
		}
		msg := prefix + strings.Join(shortages, "；")
		s.logger.Warn(fmt.Sprintf(constants.LogGardenBatchMoveRejected, userID, strings.Join(shortages, "; ")))
		return util.NewAppError(409, constants.CodeConflict, msg)
	}
	if len(changes) == 0 {
		return nil
	}

	err = s.repo.WithTx(func(tx *gorm.DB) error {
		for _, ch := range changes {
			from := ch.item.Location
			if err := tx.Model(&model.UserGarden{}).Where("id = ?", ch.item.ID).
				Update("location", ch.to).Error; err != nil {
				return err
			}
			record := &model.GardenMove{
				UserID: userID, GardenID: ch.item.ID, GardenNo: ch.item.GardenNo,
				Nickname: ch.item.Nickname, FromLocation: from, ToLocation: ch.to,
			}
			if err := s.moveRepo.Create(tx, record); err != nil {
				return err
			}
			s.logger.Info(fmt.Sprintf(constants.LogGardenMoveSuccess, ch.item.GardenNo, from, ch.to))
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("garden move apply: %w", err)
	}
	return nil
}

// moveChange is one effective relocation: a plant profile and its new location.
type moveChange struct {
	item model.UserGarden
	to   string
}

// planBatch nets the whole batch against current usage and returns the
// effective changes (no-op moves dropped) plus one shortage description per
// location whose capacity would be exceeded. Pure function: the all-or-nothing
// rule lives here so it can be unit-tested without a database.
func planBatch(byID map[uint]model.UserGarden, moves []dto.GardenBatchMoveItem,
	used map[string]int, capOf func(string) int) ([]moveChange, []string) {
	final := map[string]int{}
	for _, loc := range constants.ValidGardenLocations() {
		final[loc] = used[loc]
	}
	changes := make([]moveChange, 0, len(moves))
	for _, m := range moves {
		it := byID[m.GardenID]
		if it.Location == m.Location {
			continue
		}
		final[it.Location]--
		final[m.Location]++
		changes = append(changes, moveChange{item: it, to: m.Location})
	}
	var shortages []string
	for _, loc := range constants.ValidGardenLocations() {
		if final[loc] > capOf(loc) {
			shortages = append(shortages, fmt.Sprintf("%s缺少 %d 个位置（容量 %d，搬入后需 %d）",
				locationText(loc), final[loc]-capOf(loc), capOf(loc), final[loc]))
		}
	}
	return changes, shortages
}

// ListMoves returns the relocation history, optionally for one plant profile.
func (s *UserGardenService) ListMoves(userID, gardenID uint) ([]model.GardenMove, error) {
	items, err := s.moveRepo.ListByUser(userID, gardenID)
	if err != nil {
		return nil, fmt.Errorf("garden move list: %w", err)
	}
	return items, nil
}

// Remove soft-deletes a plant profile: it leaves the current list while move
// records and completed reminders stay queryable.
func (s *UserGardenService) Remove(userID, id uint) error {
	g, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return util.NewAppError(404, constants.CodeNotFound, fmt.Sprintf("UserGarden[id=%d] not found", id))
		}
		return fmt.Errorf("user garden remove find: %w", err)
	}
	if g.UserID != userID {
		return util.NewAppError(403, constants.CodeForbidden, fmt.Sprintf("UserGarden[id=%d] remove failed: not owner", id))
	}
	if err := s.repo.Delete(g.ID); err != nil {
		return fmt.Errorf("user garden remove: %w", err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogGardenRemoveSuccess, id, userID))
	return nil
}

// BindReminder associates a care reminder with a garden item.
func (s *UserGardenService) BindReminder(userID, gardenID, reminderID uint) (*model.UserGarden, error) {
	item, err := s.repo.FindByID(gardenID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(404, constants.CodeNotFound, fmt.Sprintf("UserGarden[id=%d] not found", gardenID))
		}
		return nil, fmt.Errorf("user garden bind find: %w", err)
	}
	if item.UserID != userID {
		return nil, util.NewAppError(403, constants.CodeForbidden, fmt.Sprintf("UserGarden[id=%d] bind failed: not owner", gardenID))
	}
	item.CareReminderID = reminderID
	if err := s.repo.Update(item); err != nil {
		return nil, fmt.Errorf("user garden bind update: %w", err)
	}
	return item, nil
}

// locationText renders a garden location in Chinese for user-facing messages.
func locationText(loc string) string {
	switch loc {
	case constants.GardenLocationIndoor:
		return "室内"
	case constants.GardenLocationBalcony:
		return "阳台"
	default:
		return loc
	}
}
