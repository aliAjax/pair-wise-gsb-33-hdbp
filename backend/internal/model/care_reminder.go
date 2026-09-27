package model

import "time"

// ReminderStatus values.
const (
	ReminderPending = "pending"
	ReminderDone    = "done"
	ReminderOverdue = "overdue"
)

// CareReminder is a scheduled gardening task owned by a user. When GardenID is
// set the reminder follows the plant profile: its location is always derived
// from the garden item, so moving the plant moves the reminder with it.
type CareReminder struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	UserID         uint      `gorm:"index;not null" json:"user_id"`
	PlantSpeciesID uint      `gorm:"index" json:"plant_species_id"`
	GardenID       uint      `gorm:"index" json:"garden_id"`
	TaskTitle      string    `gorm:"size:255;not null" json:"task_title"`
	RemindDate     time.Time `gorm:"type:date;index" json:"remind_date"`
	Frequency      string    `gorm:"size:32" json:"frequency"`
	Status         string    `gorm:"size:16;default:pending;index" json:"status"`
	CreatedAt      time.Time `json:"created_at"`

	// Transient read-only fields filled by the service layer from the linked
	// garden item so the list can show which plant (and where) a reminder
	// belongs to without extra queries from the frontend.
	GardenCode    string `gorm:"-" json:"garden_code"`
	PlantNickname string `gorm:"-" json:"plant_nickname"`
	LocationID    uint   `gorm:"-" json:"location_id"`
	LocationName  string `gorm:"-" json:"location_name"`
}
