package service

import (
	"strings"
	"testing"

	"github.com/gbplantwiki/gbplantwiki/internal/constants"
	"github.com/gbplantwiki/gbplantwiki/internal/dto"
	"github.com/gbplantwiki/gbplantwiki/internal/model"
)

func capOfDefaults(loc string) int {
	return constants.DefaultGardenLocationCapacity(loc)
}

func TestPlanBatchWithinCapacity(t *testing.T) {
	byID := map[uint]model.UserGarden{
		1: {ID: 1, GardenNo: "G-0001", Location: constants.GardenLocationIndoor},
		2: {ID: 2, GardenNo: "G-0002", Location: constants.GardenLocationIndoor},
	}
	used := map[string]int{constants.GardenLocationIndoor: 2, constants.GardenLocationBalcony: 0}
	moves := []dto.GardenBatchMoveItem{
		{GardenID: 1, Location: constants.GardenLocationBalcony},
		{GardenID: 2, Location: constants.GardenLocationBalcony},
	}
	changes, shortages := planBatch(byID, moves, used, capOfDefaults)
	if len(shortages) != 0 {
		t.Fatalf("expected no shortage, got %v", shortages)
	}
	if len(changes) != 2 {
		t.Fatalf("expected 2 changes, got %d", len(changes))
	}
}

func TestPlanBatchShortageReportsDeficit(t *testing.T) {
	byID := map[uint]model.UserGarden{
		1: {ID: 1, GardenNo: "G-0001", Location: constants.GardenLocationIndoor},
		2: {ID: 2, GardenNo: "G-0002", Location: constants.GardenLocationIndoor},
		3: {ID: 3, GardenNo: "G-0003", Location: constants.GardenLocationIndoor},
	}
	// Balcony already holds 7 of 8: moving 3 in needs 10, so 2 are missing.
	used := map[string]int{constants.GardenLocationIndoor: 3, constants.GardenLocationBalcony: 7}
	moves := []dto.GardenBatchMoveItem{
		{GardenID: 1, Location: constants.GardenLocationBalcony},
		{GardenID: 2, Location: constants.GardenLocationBalcony},
		{GardenID: 3, Location: constants.GardenLocationBalcony},
	}
	_, shortages := planBatch(byID, moves, used, capOfDefaults)
	if len(shortages) != 1 {
		t.Fatalf("expected 1 shortage, got %v", shortages)
	}
	if !strings.Contains(shortages[0], "缺少 2 个位置") || !strings.Contains(shortages[0], "阳台") {
		t.Fatalf("shortage should name balcony and deficit 2, got %q", shortages[0])
	}
}

func TestPlanBatchSwapWithinBatchFreesSlots(t *testing.T) {
	byID := map[uint]model.UserGarden{
		1: {ID: 1, GardenNo: "G-0001", Location: constants.GardenLocationIndoor},
		2: {ID: 2, GardenNo: "G-0002", Location: constants.GardenLocationBalcony},
	}
	// Full balcony (8/8): swapping one plant each way nets zero and must pass.
	used := map[string]int{constants.GardenLocationIndoor: 1, constants.GardenLocationBalcony: 8}
	moves := []dto.GardenBatchMoveItem{
		{GardenID: 1, Location: constants.GardenLocationBalcony},
		{GardenID: 2, Location: constants.GardenLocationIndoor},
	}
	changes, shortages := planBatch(byID, moves, used, capOfDefaults)
	if len(shortages) != 0 {
		t.Fatalf("swap inside one batch should fit, got %v", shortages)
	}
	if len(changes) != 2 {
		t.Fatalf("expected 2 changes, got %d", len(changes))
	}
}

func TestPlanBatchNoOpMovesDropped(t *testing.T) {
	byID := map[uint]model.UserGarden{
		1: {ID: 1, GardenNo: "G-0001", Location: constants.GardenLocationIndoor},
	}
	used := map[string]int{constants.GardenLocationIndoor: 1}
	moves := []dto.GardenBatchMoveItem{{GardenID: 1, Location: constants.GardenLocationIndoor}}
	changes, shortages := planBatch(byID, moves, used, capOfDefaults)
	if len(shortages) != 0 || len(changes) != 0 {
		t.Fatalf("no-op move should produce nothing, got changes=%v shortages=%v", changes, shortages)
	}
}
