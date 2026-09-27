package repository

import (
	"gorm.io/gorm"

	"github.com/gbplantwiki/gbplantwiki/internal/model"
)

// GardenMoveRepository handles persistence of garden relocation records.
type GardenMoveRepository struct {
	db *gorm.DB
}

// NewGardenMoveRepository creates a GardenMoveRepository.
func NewGardenMoveRepository(db *gorm.DB) *GardenMoveRepository {
	return &GardenMoveRepository{db: db}
}

// Create inserts a move record, optionally inside an existing transaction.
func (r *GardenMoveRepository) Create(tx *gorm.DB, m *model.GardenMove) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.Create(m).Error
}

// ListByUser returns move records of a user, newest first, optionally
// restricted to one plant profile. Removed plants keep their records.
func (r *GardenMoveRepository) ListByUser(userID, gardenID uint) ([]model.GardenMove, error) {
	var items []model.GardenMove
	q := r.db.Where("user_id = ?", userID)
	if gardenID > 0 {
		q = q.Where("garden_id = ?", gardenID)
	}
	if err := q.Order("id DESC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}
