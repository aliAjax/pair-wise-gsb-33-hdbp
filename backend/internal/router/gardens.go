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
	gardens.GET("/locations", h.Locations)
	gardens.PUT("/locations/:location", h.UpdateCapacity)
	gardens.GET("/moves", h.Moves)
	gardens.POST("/moves", limiter.Limit(), h.BatchMove)
	gardens.PUT("/:id/reminder", h.BindReminder)
	gardens.PUT("/:id/location", h.Move)
	gardens.DELETE("/:id", h.Remove)
}
