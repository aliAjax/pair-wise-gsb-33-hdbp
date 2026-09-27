package model

import "time"

// ReminderStatus values.
const (
	ReminderPending = "pending"
	ReminderDone    = "done"
	ReminderOverdue = "overdue"
)

// CareReminder is a scheduled gardening task owned by a user. When GardenID
// is set the reminder belongs to that plant profile and follows it across
// relocations; the profile's current location is resolved via join, never
// copied onto the reminder.
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
}

// CareReminderWithPlant is a reminder joined with its plant profile. The
// profile columns are read-only projections (the join includes soft-deleted
// profiles so completed reminders of removed plants stay queryable).
type CareReminderWithPlant struct {
	CareReminder
	GardenNo      string `json:"garden_no,omitempty"`
	PlantNickname string `json:"plant_nickname,omitempty"`
	Location      string `json:"location,omitempty"`
}
