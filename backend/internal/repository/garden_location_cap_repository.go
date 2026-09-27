package repository

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/gbplantwiki/gbplantwiki/internal/model"
)

// GardenLocationCapRepository handles persistence of per-user location capacities.
type GardenLocationCapRepository struct {
	db *gorm.DB
}

// NewGardenLocationCapRepository creates a GardenLocationCapRepository.
func NewGardenLocationCapRepository(db *gorm.DB) *GardenLocationCapRepository {
	return &GardenLocationCapRepository{db: db}
}

// MapByUser returns the user's capacity overrides keyed by location.
func (r *GardenLocationCapRepository) MapByUser(userID uint) (map[string]int, error) {
	var rows []model.GardenLocationCap
	if err := r.db.Where("user_id = ?", userID).Find(&rows).Error; err != nil {
		return nil, err
	}
	caps := make(map[string]int, len(rows))
	for _, row := range rows {
		caps[row.Location] = row.Capacity
	}
	return caps, nil
}

// Upsert sets the capacity of one location for a user.
func (r *GardenLocationCapRepository) Upsert(userID uint, location string, capacity int) error {
	row := &model.GardenLocationCap{UserID: userID, Location: location, Capacity: capacity}
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "location"}},
		DoUpdates: clause.AssignmentColumns([]string{"capacity"}),
	}).Create(row).Error
}
