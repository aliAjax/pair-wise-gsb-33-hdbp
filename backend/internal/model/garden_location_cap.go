package model

// GardenLocationCap stores a user's capacity override for one garden
// location. Locations without a row fall back to the built-in defaults.
type GardenLocationCap struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	UserID   uint   `gorm:"index:idx_loccap_user_loc,unique;not null" json:"user_id"`
	Location string `gorm:"size:32;index:idx_loccap_user_loc,unique;not null" json:"location"`
	Capacity int    `gorm:"not null" json:"capacity"`
}
