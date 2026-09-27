package model

import (
	"time"

	"gorm.io/gorm"
)

// UserGarden is a plant profile: one physical plant owned by a user. Each
// profile gets a unique per-user garden number at creation, lives in exactly
// one location (indoor/balcony) and is soft-deleted on removal so that move
// records and completed reminders stay queryable.
type UserGarden struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	UserID         uint           `gorm:"index:idx_garden_user_no,unique;not null" json:"user_id"`
	GardenNo       string         `gorm:"size:32;index:idx_garden_user_no,unique;not null" json:"garden_no"`
	PlantSpeciesID uint           `gorm:"index;not null" json:"plant_species_id"`
	Nickname       string         `gorm:"size:64" json:"nickname"`
	OwnedSince     time.Time      `gorm:"type:date" json:"owned_since"`
	Location       string         `gorm:"size:32" json:"location"`
	CareReminderID uint           `json:"care_reminder_id"`
	CreatedAt      time.Time      `json:"created_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}
