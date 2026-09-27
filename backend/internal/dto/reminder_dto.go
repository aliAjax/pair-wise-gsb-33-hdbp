package dto

// ReminderCreateRequest is the payload for creating a care reminder.
// RemindDate uses JSONDate so a plain "2006-01-02" from the date picker is
// accepted instead of being rejected for not being RFC3339.
type ReminderCreateRequest struct {
	PlantSpeciesID uint     `json:"plant_species_id"`
	GardenID       uint     `json:"garden_id"`
	TaskTitle      string   `json:"task_title" binding:"required,max=255"`
	RemindDate     JSONDate `json:"remind_date" binding:"required"`
	Frequency      string   `json:"frequency" binding:"omitempty,max=32"`
}

// ReminderStatusRequest carries the new status for a reminder.
type ReminderStatusRequest struct {
	Status string `json:"status" binding:"required"`
}
