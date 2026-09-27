package dto

import (
	"encoding/json"
	"testing"
	"time"
)

func TestJSONDateUnmarshalFormats(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"plain date", `"2026-09-27"`, "2026-09-27"},
		{"datetime", `"2026-09-27 08:30:00"`, "2026-09-27"},
		{"iso no zone", `"2026-09-27T08:30:00"`, "2026-09-27"},
		{"rfc3339", `"2026-09-27T08:30:00+08:00"`, "2026-09-27"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var d JSONDate
			if err := json.Unmarshal([]byte(tc.input), &d); err != nil {
				t.Fatalf("unmarshal %s: %v", tc.input, err)
			}
			if got := time.Time(d).Format("2006-01-02"); got != tc.want {
				t.Fatalf("got %s want %s", got, tc.want)
			}
		})
	}
}

func TestJSONDateEmptyAndNull(t *testing.T) {
	for _, input := range []string{`""`, `null`} {
		var d JSONDate
		if err := json.Unmarshal([]byte(input), &d); err != nil {
			t.Fatalf("unmarshal %s: %v", input, err)
		}
		if !time.Time(d).IsZero() {
			t.Fatalf("input %s should decode to zero time, got %v", input, d)
		}
	}
}

func TestJSONDateInvalid(t *testing.T) {
	var d JSONDate
	if err := json.Unmarshal([]byte(`"27/09/2026"`), &d); err == nil {
		t.Fatal("expected error for unsupported format")
	}
}

func TestJSONDateMarshal(t *testing.T) {
	d := JSONDate(time.Date(2026, 9, 27, 10, 0, 0, 0, time.Local))
	b, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(b) != `"2026-09-27"` {
		t.Fatalf("got %s", string(b))
	}
}
