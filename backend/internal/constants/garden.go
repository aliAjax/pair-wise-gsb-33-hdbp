package constants

// Garden location defaults and garden item status values. The location names,
// default capacities and status codes are referenced by seed data, services,
// log templates, formatters and the frontend garden page at the same time, so
// a wording change here ripples through every layer.
const (
	LocationIndoor  = "室内"
	LocationBalcony = "阳台"

	DefaultIndoorCapacity  = 10
	DefaultBalconyCapacity = 6
)

// Garden item status values (soft-delete state machine).
const (
	GardenStatusActive  = "active"
	GardenStatusRemoved = "removed"
)
