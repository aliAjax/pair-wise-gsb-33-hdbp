package handler

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/gbplantwiki/gbplantwiki/internal/constants"
	"github.com/gbplantwiki/gbplantwiki/internal/dto"
	"github.com/gbplantwiki/gbplantwiki/internal/middleware"
	"github.com/gbplantwiki/gbplantwiki/internal/model"
	"github.com/gbplantwiki/gbplantwiki/internal/service"
	"github.com/gbplantwiki/gbplantwiki/internal/util"
)

// UserGardenHandler exposes "my garden" endpoints.
type UserGardenHandler struct {
	svc    *service.UserGardenService
	logger *slog.Logger
}

// NewUserGardenHandler creates a UserGardenHandler.
func NewUserGardenHandler(svc *service.UserGardenService, logger *slog.Logger) *UserGardenHandler {
	return &UserGardenHandler{svc: svc, logger: logger}
}

// List handles GET /gardens?location=.
func (h *UserGardenHandler) List(c *gin.Context) {
	items, err := h.svc.List(middleware.GetUserID(c), c.Query("location"))
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(items))
}

// Add handles POST /gardens.
func (h *UserGardenHandler) Add(c *gin.Context) {
	var req dto.GardenAddRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, constants.CodeBadRequest, constants.MsgInvalidParam+": "+err.Error()))
		return
	}
	g := &model.UserGarden{
		PlantSpeciesID: req.PlantSpeciesID, Nickname: req.Nickname,
		OwnedSince: time.Time(req.OwnedSince), Location: req.Location,
	}
	created, err := h.svc.Add(middleware.GetUserID(c), g)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, dto.OK(created))
}

// Locations handles GET /gardens/locations.
func (h *UserGardenHandler) Locations(c *gin.Context) {
	stats, err := h.svc.LocationStats(middleware.GetUserID(c))
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(stats))
}

// UpdateCapacity handles PUT /gardens/locations/:location.
func (h *UserGardenHandler) UpdateCapacity(c *gin.Context) {
	var req dto.GardenCapacityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, constants.CodeBadRequest, constants.MsgInvalidParam))
		return
	}
	stat, err := h.svc.UpdateCapacity(middleware.GetUserID(c), c.Param("location"), req.Capacity)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(stat))
}

// Moves handles GET /gardens/moves?garden_id=.
func (h *UserGardenHandler) Moves(c *gin.Context) {
	gardenID, _ := strconv.ParseUint(c.DefaultQuery("garden_id", "0"), 10, 64)
	items, err := h.svc.ListMoves(middleware.GetUserID(c), uint(gardenID))
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(items))
}

// Move handles PUT /gardens/:id/location.
func (h *UserGardenHandler) Move(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, constants.CodeBadRequest, "invalid garden id"))
		return
	}
	var req dto.GardenMoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, constants.CodeBadRequest, constants.MsgInvalidParam))
		return
	}
	g, err := h.svc.Move(middleware.GetUserID(c), uint(id), req.Location)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(g))
}

// BatchMove handles POST /gardens/moves.
func (h *UserGardenHandler) BatchMove(c *gin.Context) {
	var req dto.GardenBatchMoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, constants.CodeBadRequest, constants.MsgInvalidParam))
		return
	}
	if err := h.svc.BatchMove(middleware.GetUserID(c), req.Moves); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"moved": len(req.Moves)}))
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

// Remove handles DELETE /gardens/:id.
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
