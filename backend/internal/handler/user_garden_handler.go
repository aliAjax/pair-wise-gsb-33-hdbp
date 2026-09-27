package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/gbplantwiki/gbplantwiki/internal/constants"
	"github.com/gbplantwiki/gbplantwiki/internal/dto"
	"github.com/gbplantwiki/gbplantwiki/internal/middleware"
	"github.com/gbplantwiki/gbplantwiki/internal/model"
	"github.com/gbplantwiki/gbplantwiki/internal/service"
	"github.com/gbplantwiki/gbplantwiki/internal/util"
)

// UserGardenHandler exposes "my garden" plant profile endpoints.
type UserGardenHandler struct {
	svc    *service.UserGardenService
	logger *slog.Logger
}

// NewUserGardenHandler creates a UserGardenHandler.
func NewUserGardenHandler(svc *service.UserGardenService, logger *slog.Logger) *UserGardenHandler {
	return &UserGardenHandler{svc: svc, logger: logger}
}

// List handles GET /gardens?location_id=.
func (h *UserGardenHandler) List(c *gin.Context) {
	locationID, _ := strconv.ParseUint(c.DefaultQuery("location_id", "0"), 10, 64)
	items, err := h.svc.List(middleware.GetUserID(c), uint(locationID))
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(items))
}

// ListRemoved handles GET /gardens/removed.
func (h *UserGardenHandler) ListRemoved(c *gin.Context) {
	items, err := h.svc.ListRemoved(middleware.GetUserID(c))
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(items))
}

// ListMoves handles GET /gardens/moves.
func (h *UserGardenHandler) ListMoves(c *gin.Context) {
	items, err := h.svc.ListMoves(middleware.GetUserID(c))
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(items))
}

// ListLocations handles GET /garden-locations.
func (h *UserGardenHandler) ListLocations(c *gin.Context) {
	views, err := h.svc.ListLocations(middleware.GetUserID(c))
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(views))
}

// UpdateLocationCapacity handles PUT /garden-locations/:id.
func (h *UserGardenHandler) UpdateLocationCapacity(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, constants.CodeBadRequest, "invalid location id"))
		return
	}
	var req dto.LocationCapacityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, constants.CodeBadRequest, constants.MsgInvalidParam))
		return
	}
	loc, err := h.svc.UpdateLocationCapacity(middleware.GetUserID(c), uint(id), req.Capacity)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(loc))
}

// Add handles POST /gardens.
func (h *UserGardenHandler) Add(c *gin.Context) {
	var req dto.GardenAddRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, constants.CodeBadRequest, constants.MsgInvalidParam))
		return
	}
	g := &model.UserGarden{
		PlantSpeciesID: req.PlantSpeciesID, Nickname: req.Nickname,
		OwnedSince: req.OwnedSince.Time(), LocationID: req.LocationID,
	}
	created, err := h.svc.Add(middleware.GetUserID(c), g)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, dto.OK(created))
}

// Move handles POST /gardens/move. The batch is atomic: if any target
// location lacks free slots the whole batch is rejected with the shortage.
func (h *UserGardenHandler) Move(c *gin.Context) {
	var req dto.GardenMoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, constants.CodeBadRequest, constants.MsgInvalidParam))
		return
	}
	items := make([]service.MoveItem, 0, len(req.Moves))
	for _, m := range req.Moves {
		items = append(items, service.MoveItem{GardenID: m.GardenID, ToLocationID: m.ToLocationID})
	}
	moved, err := h.svc.Move(middleware.GetUserID(c), items)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(moved))
}

// BindReminder handles PUT /gardens/:id/reminder.
func (h *UserGardenHandler) BindReminder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, constants.CodeBadRequest, "invalid garden id"))
		return
	}
	var req dto.GardenBindRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, constants.CodeBadRequest, constants.MsgInvalidParam))
		return
	}
	g, err := h.svc.BindReminder(middleware.GetUserID(c), uint(id), req.ReminderID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(g))
}

// Remove handles DELETE /gardens/:id. The item only leaves the current list;
// its move records and completed reminders stay queryable.
func (h *UserGardenHandler) Remove(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, constants.CodeBadRequest, "invalid garden id"))
		return
	}
	if err := h.svc.Remove(middleware.GetUserID(c), uint(id)); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"removed": true}))
}
