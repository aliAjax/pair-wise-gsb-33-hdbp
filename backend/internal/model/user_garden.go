package model

import "time"

// UserGarden represents a plant profile owned by a user inside their garden
// list. Each profile gets a unique garden code at creation and is bound to a
// GardenLocation with capacity. Removal is a soft delete: the item leaves the
// current list but its move records and completed reminders stay queryable.
type UserGarden struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	UserID         uint       `gorm:"index:idx_garden_user_plant,unique;not null" json:"user_id"`
	PlantSpeciesID uint       `gorm:"index:idx_garden_user_plant,unique;not null" json:"plant_species_id"`
	GardenCode     string     `gorm:"size:32;index" json:"garden_code"`
	Nickname       string     `gorm:"size:64" json:"nickname"`
	OwnedSince     time.Time  `gorm:"type:date" json:"owned_since"`
	LocationID     uint       `gorm:"index" json:"location_id"`
	Status         string     `gorm:"size:16;default:active;index" json:"status"`
	RemovedAt      *time.Time `json:"removed_at"`
	CareReminderID uint       `json:"care_reminder_id"`
	CreatedAt      time.Time  `json:"created_at"`

	// LocationName is a transient read-only field filled by the service layer.
	LocationName string `gorm:"-" json:"location_name"`
}
