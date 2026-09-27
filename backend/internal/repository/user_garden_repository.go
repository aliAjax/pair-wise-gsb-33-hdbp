package repository

import (
	"errors"

	"gorm.io/gorm"

	"github.com/gbplantwiki/gbplantwiki/internal/constants"
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

// CreateTx inserts a garden item inside a transaction.
func (r *UserGardenRepository) CreateTx(tx *gorm.DB, g *model.UserGarden) error {
	if err := tx.Create(g).Error; err != nil {
		if isDuplicate(err) {
			return ErrDuplicate
		}
		return err
	}
	return nil
}

// FindByID locates a garden item by primary key.
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

// FindAnyByPlant locates a garden item by user and plant regardless of status
// (active or removed), so a removed profile can be revived on re-add.
func (r *UserGardenRepository) FindAnyByPlant(userID, plantID uint) (*model.UserGarden, error) {
	var g model.UserGarden
	if err := r.db.Where("user_id = ? AND plant_species_id = ?", userID, plantID).First(&g).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &g, nil
}

// FindByIDs returns garden items of a user by primary keys, any status.
func (r *UserGardenRepository) FindByIDs(userID uint, ids []uint) ([]model.UserGarden, error) {
	var items []model.UserGarden
	if len(ids) == 0 {
		return items, nil
	}
	if err := r.db.Where("user_id = ? AND id IN ?", userID, ids).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// Update persists a garden item.
func (r *UserGardenRepository) Update(g *model.UserGarden) error {
	return r.db.Save(g).Error
}

// UpdateTx persists a garden item inside a transaction.
func (r *UserGardenRepository) UpdateTx(tx *gorm.DB, g *model.UserGarden) error {
	return tx.Save(g).Error
}

// ListActive returns active garden items of a user, optionally narrowed to a
// single location.
func (r *UserGardenRepository) ListActive(userID, locationID uint) ([]model.UserGarden, error) {
	var items []model.UserGarden
	q := r.db.Where("user_id = ? AND status = ?", userID, constants.GardenStatusActive)
	if locationID > 0 {
		q = q.Where("location_id = ?", locationID)
	}
	if err := q.Order("id DESC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// ListRemoved returns garden items that left the current list (soft-deleted).
func (r *UserGardenRepository) ListRemoved(userID uint) ([]model.UserGarden, error) {
	var items []model.UserGarden
	if err := r.db.Where("user_id = ? AND status = ?", userID, constants.GardenStatusRemoved).
		Order("removed_at DESC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// CountActiveByLocation returns how many active items sit in one location.
func (r *UserGardenRepository) CountActiveByLocation(userID, locationID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&model.UserGarden{}).
		Where("user_id = ? AND location_id = ? AND status = ?", userID, locationID, constants.GardenStatusActive).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// CountActiveGrouped returns active item counts per location for a user.
func (r *UserGardenRepository) CountActiveGrouped(userID uint) (map[uint]int64, error) {
	type row struct {
		LocationID uint
		Cnt        int64
	}
	var rows []row
	if err := r.db.Model(&model.UserGarden{}).
		Select("location_id, COUNT(*) AS cnt").
		Where("user_id = ? AND status = ?", userID, constants.GardenStatusActive).
		Group("location_id").Scan(&rows).Error; err != nil {
		return nil, err
	}
	usage := make(map[uint]int64, len(rows))
	for _, r := range rows {
		usage[r.LocationID] = r.Cnt
	}
	return usage, nil
}
