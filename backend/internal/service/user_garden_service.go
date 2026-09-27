package service

import (
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/gbplantwiki/gbplantwiki/internal/constants"
	"github.com/gbplantwiki/gbplantwiki/internal/dto"
	"github.com/gbplantwiki/gbplantwiki/internal/model"
	"github.com/gbplantwiki/gbplantwiki/internal/repository"
	"github.com/gbplantwiki/gbplantwiki/internal/util"
)

// MoveItem describes one plant relocation inside a batch move.
type MoveItem struct {
	GardenID     uint
	ToLocationID uint
}

// UserGardenService implements "my garden" plant profile logic: unique garden
// codes, location capacity checks, atomic batch moves and soft removal.
type UserGardenService struct {
	db           *gorm.DB
	repo         *repository.UserGardenRepository
	locationRepo *repository.GardenLocationRepository
	moveRepo     *repository.GardenMoveRecordRepository
	logger       *slog.Logger
}

// NewUserGardenService creates a UserGardenService.
func NewUserGardenService(db *gorm.DB, repo *repository.UserGardenRepository, locationRepo *repository.GardenLocationRepository, moveRepo *repository.GardenMoveRecordRepository, logger *slog.Logger) *UserGardenService {
	return &UserGardenService{db: db, repo: repo, locationRepo: locationRepo, moveRepo: moveRepo, logger: logger}
}

// EnsureDefaultLocations lazily creates the 室内 / 阳台 locations for a user.
func (s *UserGardenService) EnsureDefaultLocations(userID uint) error {
	locs, err := s.locationRepo.ListByUser(userID)
	if err != nil {
		return fmt.Errorf("garden location list: %w", err)
	}
	if len(locs) > 0 {
		return nil
	}
	defaults := []model.GardenLocation{
		{UserID: userID, Name: constants.LocationIndoor, Capacity: constants.DefaultIndoorCapacity},
		{UserID: userID, Name: constants.LocationBalcony, Capacity: constants.DefaultBalconyCapacity},
	}
	if err := s.locationRepo.CreateBatch(defaults); err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return nil // concurrently seeded by another request
		}
		return fmt.Errorf("garden location seed: %w", err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogGardenLocationSeeded, userID))
	return nil
}

// Add creates a plant profile in the user's garden. A unique garden code is
// generated at creation; the target location must have a free slot. If the
// same species was removed before, its profile is revived with a new location.
func (s *UserGardenService) Add(userID uint, g *model.UserGarden) (*model.UserGarden, error) {
	if err := s.EnsureDefaultLocations(userID); err != nil {
		return nil, err
	}
	g.UserID = userID
	if g.OwnedSince.IsZero() {
		g.OwnedSince = time.Now()
	}
	g.Status = constants.GardenStatusActive
	if g.LocationID == 0 {
		loc, err := s.locationRepo.FindByName(userID, constants.LocationIndoor)
		if err != nil {
			return nil, fmt.Errorf("garden add default location: %w", err)
		}
		g.LocationID = loc.ID
	}
	if err := s.checkCapacity(userID, g.LocationID, 1); err != nil {
		return nil, err
	}

	existing, err := s.repo.FindAnyByPlant(userID, g.PlantSpeciesID)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, fmt.Errorf("user garden add find: %w", err)
	}
	if err == nil {
		if existing.Status == constants.GardenStatusActive {
			return nil, util.NewAppError(409, constants.CodeConflict,
				fmt.Sprintf("UserGarden[user_id=%d plant_id=%d] add failed: already in garden", userID, g.PlantSpeciesID))
		}
		// Revive the removed profile: the garden code stays the same and the
		// old move records keep pointing at this plant.
		existing.Status = constants.GardenStatusActive
		existing.RemovedAt = nil
		existing.LocationID = g.LocationID
		existing.OwnedSince = g.OwnedSince
		if g.Nickname != "" {
			existing.Nickname = g.Nickname
		}
		if err := s.repo.Update(existing); err != nil {
			return nil, fmt.Errorf("user garden revive: %w", err)
		}
		s.logger.Info(fmt.Sprintf(constants.LogGardenAddSuccess, existing.PlantSpeciesID, userID, existing.GardenCode), "id", existing.ID)
		existing.LocationName = s.locationNameOf(existing.LocationID)
		return existing, nil
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.CreateTx(tx, g); err != nil {
			return err
		}
		g.GardenCode = util.FormatGardenCode(g.ID)
		return s.repo.UpdateTx(tx, g)
	})
	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return nil, util.NewAppError(409, constants.CodeConflict,
				fmt.Sprintf("UserGarden[user_id=%d plant_id=%d] add failed: already in garden", userID, g.PlantSpeciesID))
		}
		s.logger.Error(fmt.Sprintf(constants.LogGardenAddFailed, g.PlantSpeciesID, userID), "error", err)
		return nil, fmt.Errorf("user garden add: %w", err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogGardenAddSuccess, g.PlantSpeciesID, userID, g.GardenCode), "id", g.ID)
	g.LocationName = s.locationNameOf(g.LocationID)
	return g, nil
}

// List returns a user's active garden items, optionally filtered by location,
// with the location name filled in.
func (s *UserGardenService) List(userID, locationID uint) ([]model.UserGarden, error) {
	if err := s.EnsureDefaultLocations(userID); err != nil {
		return nil, err
	}
	items, err := s.repo.ListActive(userID, locationID)
	if err != nil {
		return nil, fmt.Errorf("user garden list: %w", err)
	}
	if err := s.fillLocationNames(userID, items); err != nil {
		return nil, err
	}
	return items, nil
}

// ListRemoved returns garden items that left the current list.
func (s *UserGardenService) ListRemoved(userID uint) ([]model.UserGarden, error) {
	items, err := s.repo.ListRemoved(userID)
	if err != nil {
		return nil, fmt.Errorf("user garden removed list: %w", err)
	}
	if err := s.fillLocationNames(userID, items); err != nil {
		return nil, err
	}
	return items, nil
}

// ListLocations returns the user's locations with used/remaining slots.
func (s *UserGardenService) ListLocations(userID uint) ([]dto.LocationView, error) {
	if err := s.EnsureDefaultLocations(userID); err != nil {
		return nil, err
	}
	locs, err := s.locationRepo.ListByUser(userID)
	if err != nil {
		return nil, fmt.Errorf("garden location list: %w", err)
	}
	usage, err := s.repo.CountActiveGrouped(userID)
	if err != nil {
		return nil, fmt.Errorf("garden location usage: %w", err)
	}
	views := make([]dto.LocationView, 0, len(locs))
	for _, loc := range locs {
		used := usage[loc.ID]
		views = append(views, dto.LocationView{
			ID:        loc.ID,
			UserID:    loc.UserID,
			Name:      loc.Name,
			Capacity:  loc.Capacity,
			Used:      used,
			Remaining: int64(loc.Capacity) - used,
			CreatedAt: util.FormatDateTime(loc.CreatedAt),
		})
	}
	return views, nil
}

// UpdateLocationCapacity changes a location's capacity. The new capacity must
// hold the plants currently sitting there.
func (s *UserGardenService) UpdateLocationCapacity(userID, locationID uint, capacity int) (*model.GardenLocation, error) {
	loc, err := s.locationRepo.FindByID(locationID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(404, constants.CodeNotFound, fmt.Sprintf("GardenLocation[id=%d] not found", locationID))
		}
		return nil, fmt.Errorf("garden location find: %w", err)
	}
	if loc.UserID != userID {
		return nil, util.NewAppError(403, constants.CodeForbidden,
			fmt.Sprintf("GardenLocation[id=%d] capacity update failed: user_id=%d not owner", locationID, userID))
	}
	used, err := s.repo.CountActiveByLocation(userID, locationID)
	if err != nil {
		return nil, fmt.Errorf("garden location usage: %w", err)
	}
	if int64(capacity) < used {
		return nil, util.NewAppError(422, constants.CodeValidationError,
			fmt.Sprintf("GardenLocation[id=%d name=%s] capacity=%d invalid: %d plant(s) currently placed", locationID, loc.Name, capacity, used))
	}
	loc.Capacity = capacity
	if err := s.locationRepo.Update(loc); err != nil {
		return nil, fmt.Errorf("garden location update: %w", err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogLocationCapacityUpdated, loc.ID, loc.Name, loc.Capacity))
	return loc, nil
}

// Move relocates plants in a single atomic batch. Every target location is
// checked for remaining slots first; if any location cannot fit the incoming
// plants the whole batch is rejected and the error states how many slots are
// missing. Each applied move writes an immutable move record.
func (s *UserGardenService) Move(userID uint, items []MoveItem) ([]model.UserGarden, error) {
	if err := s.EnsureDefaultLocations(userID); err != nil {
		return nil, err
	}
	seen := make(map[uint]bool, len(items))
	ids := make([]uint, 0, len(items))
	for _, it := range items {
		if seen[it.GardenID] {
			return nil, util.NewAppError(400, constants.CodeBadRequest,
				fmt.Sprintf("UserGarden[id=%d] move failed: duplicate garden_id in batch", it.GardenID))
		}
		seen[it.GardenID] = true
		ids = append(ids, it.GardenID)
	}

	gardens, err := s.repo.FindByIDs(userID, ids)
	if err != nil {
		return nil, fmt.Errorf("user garden move find: %w", err)
	}
	gardenMap := make(map[uint]*model.UserGarden, len(gardens))
	for i := range gardens {
		gardenMap[gardens[i].ID] = &gardens[i]
	}
	fromByGarden := make(map[uint]uint, len(items))
	for _, it := range items {
		g, ok := gardenMap[it.GardenID]
		if !ok {
			return nil, util.NewAppError(404, constants.CodeNotFound, fmt.Sprintf("UserGarden[id=%d] not found", it.GardenID))
		}
		if g.Status != constants.GardenStatusActive {
			return nil, util.NewAppError(422, constants.CodeValidationError,
				fmt.Sprintf("UserGarden[id=%d code=%s] move failed: status=%s not active", g.ID, g.GardenCode, g.Status))
		}
		fromByGarden[it.GardenID] = g.LocationID
	}

	locs, err := s.locationRepo.ListByUser(userID)
	if err != nil {
		return nil, fmt.Errorf("garden location list: %w", err)
	}
	locMap := make(map[uint]model.GardenLocation, len(locs))
	for _, loc := range locs {
		locMap[loc.ID] = loc
	}
	for _, it := range items {
		if _, ok := locMap[it.ToLocationID]; !ok {
			return nil, util.NewAppError(404, constants.CodeNotFound,
				fmt.Sprintf("GardenLocation[id=%d] move failed: target not found", it.ToLocationID))
		}
	}

	usage, err := s.repo.CountActiveGrouped(userID)
	if err != nil {
		return nil, fmt.Errorf("garden location usage: %w", err)
	}
	if shorts := CheckBatchMove(usage, locMap, fromByGarden, items); len(shorts) > 0 {
		reason := strings.Join(shorts, "；")
		s.logger.Warn(fmt.Sprintf(constants.LogGardenMoveRejected, userID, reason))
		return nil, util.NewAppError(409, constants.CodeLocationFull,
			fmt.Sprintf("UserGarden[user_id=%d] batch move rejected: %s", userID, reason))
	}

	now := time.Now()
	err = s.db.Transaction(func(tx *gorm.DB) error {
		for _, it := range items {
			g := gardenMap[it.GardenID]
			if g.LocationID == it.ToLocationID {
				continue
			}
			rec := &model.GardenMoveRecord{
				UserID:        userID,
				GardenID:      g.ID,
				GardenCode:    g.GardenCode,
				PlantNickname: g.Nickname,
				FromLocation:  locationName(locMap, g.LocationID),
				ToLocation:    locationName(locMap, it.ToLocationID),
				MovedAt:       now,
			}
			if err := s.moveRepo.CreateTx(tx, rec); err != nil {
				return fmt.Errorf("garden move record: %w", err)
			}
			g.LocationID = it.ToLocationID
			if err := s.repo.UpdateTx(tx, g); err != nil {
				return fmt.Errorf("garden move update: %w", err)
			}
			s.logger.Info(fmt.Sprintf(constants.LogGardenMoveSuccess, g.ID, g.GardenCode, rec.FromLocation, rec.ToLocation))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	result := make([]model.UserGarden, 0, len(items))
	for _, it := range items {
		g := gardenMap[it.GardenID]
		g.LocationName = locationName(locMap, g.LocationID)
		result = append(result, *g)
	}
	return result, nil
}

// ListMoves returns the user's move history, including records of plants that
// have since been removed from the garden.
func (s *UserGardenService) ListMoves(userID uint) ([]model.GardenMoveRecord, error) {
	items, err := s.moveRepo.ListByUser(userID)
	if err != nil {
		return nil, fmt.Errorf("garden move list: %w", err)
	}
	return items, nil
}

// Remove soft-deletes a garden item: it leaves the current list but its move
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
		return util.NewAppError(403, constants.CodeForbidden,
			fmt.Sprintf("UserGarden[id=%d] remove failed: user_id=%d not owner", id, userID))
	}
	if g.Status == constants.GardenStatusRemoved {
		return nil
	}
	now := time.Now()
	g.Status = constants.GardenStatusRemoved
	g.RemovedAt = &now
	if err := s.repo.Update(g); err != nil {
		return fmt.Errorf("user garden remove: %w", err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogGardenRemoveSuccess, g.ID, g.GardenCode, util.GardenStatusText(g.Status)), "id", id)
	return nil
}

// BindReminder associates a care reminder with a garden item.
func (s *UserGardenService) BindReminder(userID, gardenID, reminderID uint) (*model.UserGarden, error) {
	item, err := s.repo.FindByID(gardenID)
	if err != nil {
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

// checkCapacity verifies that a location owned by the user can take extra
// more plants; the error states how many slots are missing.
func (s *UserGardenService) checkCapacity(userID, locationID uint, extra int64) error {
	loc, err := s.locationRepo.FindByID(locationID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return util.NewAppError(404, constants.CodeNotFound, fmt.Sprintf("GardenLocation[id=%d] not found", locationID))
		}
		return fmt.Errorf("garden location find: %w", err)
	}
	if loc.UserID != userID {
		return util.NewAppError(403, constants.CodeForbidden,
			fmt.Sprintf("GardenLocation[id=%d] capacity check failed: user_id=%d not owner", locationID, userID))
	}
	used, err := s.repo.CountActiveByLocation(userID, locationID)
	if err != nil {
		return fmt.Errorf("garden location usage: %w", err)
	}
	if used+extra > int64(loc.Capacity) {
		short := used + extra - int64(loc.Capacity)
		return util.NewAppError(409, constants.CodeLocationFull,
			fmt.Sprintf("GardenLocation[id=%d name=%s] 容量不足，还差 %d 个空位（容量 %d，当前占用 %d）",
				loc.ID, loc.Name, short, loc.Capacity, used))
	}
	return nil
}

// fillLocationNames populates the transient LocationName of garden items.
func (s *UserGardenService) fillLocationNames(userID uint, items []model.UserGarden) error {
	locs, err := s.locationRepo.ListByUser(userID)
	if err != nil {
		return fmt.Errorf("garden location list: %w", err)
	}
	locMap := make(map[uint]model.GardenLocation, len(locs))
	for _, loc := range locs {
		locMap[loc.ID] = loc
	}
	for i := range items {
		items[i].LocationName = locationName(locMap, items[i].LocationID)
	}
	return nil
}

// locationNameOf resolves a single location id to its name, ignoring lookup
// errors (the name is a display convenience, not critical data).
func (s *UserGardenService) locationNameOf(locationID uint) string {
	loc, err := s.locationRepo.FindByID(locationID)
	if err != nil {
		return ""
	}
	return loc.Name
}

// locationName resolves a location id to its name, or 未分配 for legacy items
// without a location.
func locationName(locMap map[uint]model.GardenLocation, id uint) string {
	if loc, ok := locMap[id]; ok {
		return loc.Name
	}
	return "未分配"
}

// CheckBatchMove simulates a batch move on the current per-location usage and
// returns one human readable shortage entry ("缺多少") per location that would
// overflow. It is a pure function so the capacity rule can be unit tested and
// reused by any caller that already loaded usage, locations and moves.
func CheckBatchMove(usage map[uint]int64, locations map[uint]model.GardenLocation, fromByGarden map[uint]uint, moves []MoveItem) []string {
	sim := make(map[uint]int64, len(usage)+len(locations))
	for k, v := range usage {
		sim[k] = v
	}
	for _, m := range moves {
		from := fromByGarden[m.GardenID]
		if from == m.ToLocationID {
			continue
		}
		sim[from]--
		sim[m.ToLocationID]++
	}
	var shorts []string
	for locID, loc := range locations {
		if sim[locID] > int64(loc.Capacity) {
			shorts = append(shorts, fmt.Sprintf("位置[%s]容量不足，还差 %d 个空位（容量 %d，当前占用 %d）",
				loc.Name, sim[locID]-int64(loc.Capacity), loc.Capacity, sim[locID]))
		}
	}
	sort.Strings(shorts)
	return shorts
}
