package dto

// ReminderCreateRequest is the payload for creating a care reminder. When
// GardenID is set the reminder is bound to that plant profile and follows it
// across locations.
type ReminderCreateRequest struct {
	PlantSpeciesID uint     `json:"plant_species_id"`
	GardenID       uint     `json:"garden_id"`
	TaskTitle      string   `json:"task_title" binding:"required,max=255"`
	RemindDate     JSONDate `json:"remind_date"`
	Frequency      string   `json:"frequency" binding:"omitempty,max=32"`
}

// ReminderStatusRequest carries the new status for a reminder.
type ReminderStatusRequest struct {
	Status string `json:"status" binding:"required"`
}
