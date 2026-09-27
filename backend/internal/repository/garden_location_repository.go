package repository

import (
	"errors"

	"gorm.io/gorm"

	"github.com/gbplantwiki/gbplantwiki/internal/model"
)

// GardenLocationRepository handles persistence of garden locations.
type GardenLocationRepository struct {
	db *gorm.DB
}

// NewGardenLocationRepository creates a GardenLocationRepository.
func NewGardenLocationRepository(db *gorm.DB) *GardenLocationRepository {
	return &GardenLocationRepository{db: db}
}

// CreateBatch inserts default locations for a user.
func (r *GardenLocationRepository) CreateBatch(locs []model.GardenLocation) error {
	if err := r.db.Create(&locs).Error; err != nil {
		if isDuplicate(err) {
			return ErrDuplicate
		}
		return err
	}
	return nil
}

// ListByUser returns all locations of a user ordered by id.
func (r *GardenLocationRepository) ListByUser(userID uint) ([]model.GardenLocation, error) {
	var items []model.GardenLocation
	if err := r.db.Where("user_id = ?", userID).Order("id ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// FindByID locates a location by primary key.
func (r *GardenLocationRepository) FindByID(id uint) (*model.GardenLocation, error) {
	var loc model.GardenLocation
	if err := r.db.First(&loc, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &loc, nil
}

// FindByName locates a location by user and name.
func (r *GardenLocationRepository) FindByName(userID uint, name string) (*model.GardenLocation, error) {
	var loc model.GardenLocation
	if err := r.db.Where("user_id = ? AND name = ?", userID, name).First(&loc).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &loc, nil
}

// Update persists a location.
func (r *GardenLocationRepository) Update(loc *model.GardenLocation) error {
	return r.db.Save(loc).Error
}
