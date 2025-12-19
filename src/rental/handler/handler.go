package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	services "github.com/SwanPoi/bmstu_rsoi_lab2/src/rental/services"
)

type RentalHandler struct {
	services *services.Services
}

func NewHandler(services *services.Services) *RentalHandler {
	return &RentalHandler{services: services}
}

func (h *RentalHandler) SetupRoutes() *gin.Engine {
	router := gin.New()

	router.GET("/manage/health", func (c *gin.Context) {
		c.Status(http.StatusOK)
	})

	api := router.Group("/api/v1") 
	{
		rentals := api.Group("/rental")
		{
			rentals.GET("", h.GetUserRentals)
			rentals.GET("/:uid", h.GetUserRentalByUid)
			rentals.POST("", h.CreateRental)
			rentals.PATCH("/:uid", h.UpdateRental)
		}
	}

	return router
}