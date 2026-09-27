package constants

// GardenLocation enumerates the fixed placement areas inside a user's garden.
// Every plant profile lives in exactly one location and each location has a
// capacity; the same values are mirrored in the frontend filter and labels.
const (
	GardenLocationIndoor  = "indoor"  // 室内
	GardenLocationBalcony = "balcony" // 阳台
)

// Default capacities per garden location, used when the user has no override.
const (
	DefaultIndoorCapacity  = 12
	DefaultBalconyCapacity = 8
)

// ValidGardenLocations returns all accepted garden location values.
func ValidGardenLocations() []string {
	return []string{GardenLocationIndoor, GardenLocationBalcony}
}

// IsValidGardenLocation reports whether the given location is a known garden location.
func IsValidGardenLocation(loc string) bool {
	for _, v := range ValidGardenLocations() {
		if v == loc {
			return true
		}
	}
	return false
}

// DefaultGardenLocationCapacity returns the built-in capacity of a location.
func DefaultGardenLocationCapacity(loc string) int {
	switch loc {
	case GardenLocationIndoor:
		return DefaultIndoorCapacity
	case GardenLocationBalcony:
		return DefaultBalconyCapacity
	default:
		return 0
	}
}
