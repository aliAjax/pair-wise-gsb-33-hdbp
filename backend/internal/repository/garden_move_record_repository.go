package repository

import (
	"gorm.io/gorm"

	"github.com/gbplantwiki/gbplantwiki/internal/model"
)

// GardenMoveRecordRepository handles persistence of garden move history.
type GardenMoveRecordRepository struct {
	db *gorm.DB
}

// NewGardenMoveRecordRepository creates a GardenMoveRecordRepository.
func NewGardenMoveRecordRepository(db *gorm.DB) *GardenMoveRecordRepository {
	return &GardenMoveRecordRepository{db: db}
}

// CreateTx inserts a move record inside a transaction.
func (r *GardenMoveRecordRepository) CreateTx(tx *gorm.DB, rec *model.GardenMoveRecord) error {
	return tx.Create(rec).Error
}

// ListByUser returns all move records of a user, newest first. Records of
// removed garden items are included so history stays queryable.
func (r *GardenMoveRecordRepository) ListByUser(userID uint) ([]model.GardenMoveRecord, error) {
	var items []model.GardenMoveRecord
	if err := r.db.Where("user_id = ?", userID).Order("id DESC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}
