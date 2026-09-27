package router

import (
	"github.com/gin-gonic/gin"

	"github.com/gbplantwiki/gbplantwiki/internal/config"
	"github.com/gbplantwiki/gbplantwiki/internal/handler"
	"github.com/gbplantwiki/gbplantwiki/internal/middleware"
)

func registerGardenRoutes(v1 *gin.RouterGroup, cfg *config.Config, h *handler.UserGardenHandler, limiter *middleware.RateLimiter) {
	gardens := v1.Group("/gardens", middleware.AuthRequired(cfg))
	gardens.GET("", h.List)
	gardens.POST("", limiter.Limit(), h.Add)
	gardens.GET("/removed", h.ListRemoved)
	gardens.GET("/moves", h.ListMoves)
	gardens.POST("/move", limiter.Limit(), h.Move)
	gardens.PUT("/:id/reminder", h.BindReminder)
	gardens.DELETE("/:id", h.Remove)

	locations := v1.Group("/garden-locations", middleware.AuthRequired(cfg))
	locations.GET("", h.ListLocations)
	locations.PUT("/:id", h.UpdateLocationCapacity)
}
