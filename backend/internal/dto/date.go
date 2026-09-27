package dto

import (
	"fmt"
	"strings"
	"time"
)

// JSONDate is a lenient date type for request bodies. It accepts "2006-01-02"
// (what the frontend date picker submits), RFC3339 timestamps and a couple of
// common datetime layouts, so picking a date on the page and hitting save is
// never rejected by the API just because of the date format.
type JSONDate time.Time

var jsonDateLayouts = []string{
	"2006-01-02",
	time.RFC3339,
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05",
}

// UnmarshalJSON parses a JSON string into a JSONDate using the known layouts.
func (d *JSONDate) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	if s == "" || s == "null" {
		return nil
	}
	for _, layout := range jsonDateLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			*d = JSONDate(t)
			return nil
		}
	}
	return fmt.Errorf("unsupported date format %q", s)
}

// Time returns the underlying time.Time value.
func (d JSONDate) Time() time.Time {
	return time.Time(d)
}

// IsZero reports whether the date is unset.
func (d JSONDate) IsZero() bool {
	return time.Time(d).IsZero()
}
