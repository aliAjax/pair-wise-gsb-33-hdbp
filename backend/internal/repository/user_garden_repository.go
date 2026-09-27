package repository

import (
	"errors"

	"gorm.io/gorm"

	"github.com/gbplantwiki/gbplantwiki/internal/model"
)

// UserGardenRepository handles persistence of user garden items.
type UserGardenRepository struct {
	db *gorm.DB
}

// NewUserGardenRepository creates a UserGardenRepository.
func NewUserGardenRepository(db *gorm.DB) *UserGardenRepository {
	return &UserGardenRepository{db: db}
}

// Create inserts a garden item.
func (r *UserGardenRepository) Create(g *model.UserGarden) error {
	if err := r.db.Create(g).Error; err != nil {
		if isDuplicate(err) {
			return ErrDuplicate
		}
		return err
	}
	return nil
}

// Find locates an active garden item by user and plant species.
func (r *UserGardenRepository) Find(userID, plantID uint) (*model.UserGarden, error) {
	var g model.UserGarden
	if err := r.db.Where("user_id = ? AND plant_species_id = ?", userID, plantID).First(&g).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &g, nil
}

// FindByID locates an active garden item by primary key.
func (r *UserGardenRepository) FindByID(id uint) (*model.UserGarden, error) {
	var g model.UserGarden
	if err := r.db.First(&g, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &g, nil
}

// FindActiveByIDs returns the active garden items of a user matching the ids.
func (r *UserGardenRepository) FindActiveByIDs(userID uint, ids []uint) ([]model.UserGarden, error) {
	var items []model.UserGarden
	if err := r.db.Where("user_id = ? AND id IN ?", userID, ids).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// Update persists a garden item.
func (r *UserGardenRepository) Update(g *model.UserGarden) error {
	return r.db.Save(g).Error
}

// Delete soft-removes a garden item by id: the row stays so move records and
// completed reminders keep their references.
func (r *UserGardenRepository) Delete(id uint) error {
	return r.db.Delete(&model.UserGarden{}, id).Error
}

// ListByUser returns active garden items of a user, optionally filtered by location.
func (r *UserGardenRepository) ListByUser(userID uint, location string) ([]model.UserGarden, error) {
	var items []model.UserGarden
	q := r.db.Where("user_id = ?", userID)
	if location != "" {
		q = q.Where("location = ?", location)
	}
	if err := q.Order("id DESC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// CountByLocation counts active garden items of a user grouped by location.
func (r *UserGardenRepository) CountByLocation(userID uint) (map[string]int, error) {
	type row struct {
		Location string
		Cnt      int
	}
	var rows []row
	if err := r.db.Model(&model.UserGarden{}).
		Select("location, COUNT(*) AS cnt").
		Where("user_id = ?", userID).
		Group("location").Scan(&rows).Error; err != nil {
		return nil, err
	}
	counts := make(map[string]int, len(rows))
	for _, r := range rows {
		counts[r.Location] = r.Cnt
	}
	return counts, nil
}

// CountAllByUser counts every garden item of a user including soft-deleted
// ones, so garden numbers are never reused.
func (r *UserGardenRepository) CountAllByUser(userID uint) (int64, error) {
	var total int64
	if err := r.db.Unscoped().Model(&model.UserGarden{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// WithTx runs fn inside a transaction, committing on success.
func (r *UserGardenRepository) WithTx(fn func(tx *gorm.DB) error) error {
	return r.db.Transaction(fn)
}
