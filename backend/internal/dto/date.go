package dto

import (
	"fmt"
	"strings"
	"time"
)

// JSONDate is a request-side date that accepts the formats the frontend can
// produce: "2006-01-02" (date picker value-format), "2006-01-02 15:04:05" and
// RFC3339. Empty string and null decode to the zero time.
//
// It is defined as a time.Time alias (not a wrapper struct) on purpose: the
// validator skips `required` on non-pointer struct fields unless they are
// convertible to time.Time, so the alias keeps `binding:"required"` working.
type JSONDate time.Time

// UnmarshalJSON decodes a date string in any supported layout.
func (d *JSONDate) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "" || s == "null" {
		*d = JSONDate(time.Time{})
		return nil
	}
	layouts := []string{"2006-01-02", "2006-01-02 15:04:05", "2006-01-02T15:04:05", time.RFC3339}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			*d = JSONDate(t)
			return nil
		}
	}
	return fmt.Errorf("unsupported date format: %q", s)
}

// MarshalJSON renders the date as "2006-01-02" (empty when zero).
func (d JSONDate) MarshalJSON() ([]byte, error) {
	t := time.Time(d)
	if t.IsZero() {
		return []byte(`""`), nil
	}
	return []byte(`"` + t.Format("2006-01-02") + `"`), nil
}
