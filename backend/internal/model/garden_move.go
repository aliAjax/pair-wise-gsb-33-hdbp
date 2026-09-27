package model

import "time"

// GardenMove records one relocation of a plant profile between garden
// locations. Records are kept forever, even after the plant is removed from
// the current list, so the relocation history stays queryable.
type GardenMove struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"index:idx_move_user_garden;not null" json:"user_id"`
	GardenID     uint      `gorm:"index:idx_move_user_garden;not null" json:"garden_id"`
	GardenNo     string    `gorm:"size:32" json:"garden_no"`
	Nickname     string    `gorm:"size:64" json:"nickname"`
	FromLocation string    `gorm:"size:32" json:"from_location"`
	ToLocation   string    `gorm:"size:32" json:"to_location"`
	CreatedAt    time.Time `json:"created_at"`
}
