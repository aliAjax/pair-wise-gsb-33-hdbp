package service

import (
	"testing"

	"github.com/gbplantwiki/gbplantwiki/internal/model"
)

func TestCheckBatchMove(t *testing.T) {
	locations := map[uint]model.GardenLocation{
		1: {ID: 1, Name: "室内", Capacity: 2},
		2: {ID: 2, Name: "阳台", Capacity: 1},
	}

	tests := []struct {
		name         string
		usage        map[uint]int64
		fromByGarden map[uint]uint
		moves        []MoveItem
		wantShorts   int
	}{
		{
			name:         "fits within remaining capacity",
			usage:        map[uint]int64{1: 1, 2: 0},
			fromByGarden: map[uint]uint{10: 1},
			moves:        []MoveItem{{GardenID: 10, ToLocationID: 2}},
			wantShorts:   0,
		},
		{
			name:         "target full rejects with shortage",
			usage:        map[uint]int64{1: 1, 2: 1},
			fromByGarden: map[uint]uint{10: 1},
			moves:        []MoveItem{{GardenID: 10, ToLocationID: 2}},
			wantShorts:   1,
		},
		{
			name:         "swap inside batch nets zero",
			usage:        map[uint]int64{1: 2, 2: 1},
			fromByGarden: map[uint]uint{10: 1, 11: 2},
			moves: []MoveItem{
				{GardenID: 10, ToLocationID: 2},
				{GardenID: 11, ToLocationID: 1},
			},
			wantShorts: 0,
		},
		{
			name:         "any overflow rejects the whole batch and reports each shortage",
			usage:        map[uint]int64{1: 2, 2: 1},
			fromByGarden: map[uint]uint{10: 0, 11: 0, 12: 0},
			moves: []MoveItem{
				{GardenID: 10, ToLocationID: 1},
				{GardenID: 11, ToLocationID: 2},
				{GardenID: 12, ToLocationID: 2},
			},
			wantShorts: 2,
		},
		{
			name:         "move to same location is a no-op",
			usage:        map[uint]int64{1: 2},
			fromByGarden: map[uint]uint{10: 1},
			moves:        []MoveItem{{GardenID: 10, ToLocationID: 1}},
			wantShorts:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shorts := CheckBatchMove(tt.usage, locations, tt.fromByGarden, tt.moves)
			if len(shorts) != tt.wantShorts {
				t.Fatalf("expected %d shortage entries, got %d: %v", tt.wantShorts, len(shorts), shorts)
			}
			// usage maps must not be mutated by the simulation
			if tt.name == "target full rejects with shortage" && tt.usage[2] != 1 {
				t.Errorf("usage map mutated: %v", tt.usage)
			}
		})
	}
}

func TestCheckBatchMoveShortageMessage(t *testing.T) {
	locations := map[uint]model.GardenLocation{
		2: {ID: 2, Name: "阳台", Capacity: 1},
	}
	usage := map[uint]int64{2: 1}
	fromByGarden := map[uint]uint{10: 0, 11: 0}
	moves := []MoveItem{{GardenID: 10, ToLocationID: 2}, {GardenID: 11, ToLocationID: 2}}

	shorts := CheckBatchMove(usage, locations, fromByGarden, moves)
	if len(shorts) != 1 {
		t.Fatalf("expected 1 shortage entry, got %v", shorts)
	}
	// 1 free slot, 2 incoming -> must state that 2 slots are missing.
	want := "位置[阳台]容量不足，还差 2 个空位（容量 1，当前占用 3）"
	if shorts[0] != want {
		t.Errorf("shortage message mismatch:\n got: %s\nwant: %s", shorts[0], want)
	}
}
