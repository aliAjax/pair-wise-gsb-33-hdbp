package dto

// GardenAddRequest adds a plant to the user's garden as a plant profile.
type GardenAddRequest struct {
	PlantSpeciesID uint     `json:"plant_species_id" binding:"required"`
	Nickname       string   `json:"nickname" binding:"omitempty,max=64"`
	OwnedSince     JSONDate `json:"owned_since"`
	LocationID     uint     `json:"location_id"`
}

// GardenBindRequest binds a reminder to a garden item.
type GardenBindRequest struct {
	ReminderID uint `json:"care_reminder_id" binding:"required"`
}

// GardenMoveItem describes one plant relocation inside a batch move.
type GardenMoveItem struct {
	GardenID     uint `json:"garden_id" binding:"required"`
	ToLocationID uint `json:"to_location_id" binding:"required"`
}

// GardenMoveRequest is the batch move payload. The whole batch is rejected if
// any target location cannot fit the incoming plants.
type GardenMoveRequest struct {
	Moves []GardenMoveItem `json:"moves" binding:"required,min=1,dive"`
}

// LocationCapacityRequest updates the capacity of a garden location.
type LocationCapacityRequest struct {
	Capacity int `json:"capacity" binding:"required,min=0"`
}

// LocationView is the garden location response enriched with usage numbers.
type LocationView struct {
	ID        uint   `json:"id"`
	UserID    uint   `json:"user_id"`
	Name      string `json:"name"`
	Capacity  int    `json:"capacity"`
	Used      int64  `json:"used"`
	Remaining int64  `json:"remaining"`
	CreatedAt string `json:"created_at"`
}
