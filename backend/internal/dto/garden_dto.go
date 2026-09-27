package dto

// GardenAddRequest creates a plant profile in the user's garden.
type GardenAddRequest struct {
	PlantSpeciesID uint     `json:"plant_species_id" binding:"required"`
	Nickname       string   `json:"nickname" binding:"omitempty,max=64"`
	OwnedSince     JSONDate `json:"owned_since"`
	Location       string   `json:"location" binding:"omitempty,max=32"`
}

// GardenBindRequest binds a reminder to a garden item.
type GardenBindRequest struct {
	ReminderID uint `json:"care_reminder_id" binding:"required"`
}

// GardenMoveRequest relocates one plant profile to another location.
type GardenMoveRequest struct {
	Location string `json:"location" binding:"required,max=32"`
}

// GardenBatchMoveItem is one relocation inside a batch move.
type GardenBatchMoveItem struct {
	GardenID uint   `json:"garden_id" binding:"required"`
	Location string `json:"location" binding:"required,max=32"`
}

// GardenBatchMoveRequest relocates several plant profiles atomically: if any
// target location lacks capacity the whole batch is rejected.
type GardenBatchMoveRequest struct {
	Moves []GardenBatchMoveItem `json:"moves" binding:"required,min=1,dive"`
}

// GardenCapacityRequest sets the capacity of one garden location.
type GardenCapacityRequest struct {
	Capacity int `json:"capacity" binding:"required,gt=0"`
}

// GardenLocationStat reports capacity usage of one garden location.
type GardenLocationStat struct {
	Location  string `json:"location"`
	Capacity  int    `json:"capacity"`
	Used      int    `json:"used"`
	Remaining int    `json:"remaining"`
}
