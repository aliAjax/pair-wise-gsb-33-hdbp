package model

import "time"

// GardenLocation is a named planting spot (室内 / 阳台) owned by a user with a
// fixed capacity. Moves and additions are checked against the remaining slots.
type GardenLocation struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index:idx_location_user_name,unique;not null" json:"user_id"`
	Name      string    `gorm:"size:32;index:idx_location_user_name,unique;not null" json:"name"`
	Capacity  int       `gorm:"not null;default:0" json:"capacity"`
	CreatedAt time.Time `json:"created_at"`
}
