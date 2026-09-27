package model

import "time"

// GardenMoveRecord is an immutable history entry for a plant relocation.
// Records survive the (soft) removal of the garden item they reference so old
// moves stay queryable after a plant leaves the current list.
type GardenMoveRecord struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uint      `gorm:"index;not null" json:"user_id"`
	GardenID      uint      `gorm:"index;not null" json:"garden_id"`
	GardenCode    string    `gorm:"size:32" json:"garden_code"`
	PlantNickname string    `gorm:"size:64" json:"plant_nickname"`
	FromLocation  string    `gorm:"size:32" json:"from_location"`
	ToLocation    string    `gorm:"size:32" json:"to_location"`
	MovedAt       time.Time `json:"moved_at"`
}
