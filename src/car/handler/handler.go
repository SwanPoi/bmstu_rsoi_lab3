package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	services "github.com/SwanPoi/bmstu_rsoi_lab2/src/car/services"
)

type CarHandler struct {
	services *services.Services
}

func NewHandler(services *services.Services) *CarHandler {
	return &CarHandler{services: services}
}

func (h *CarHandler) SetupRoutes() *gin.Engine {
	router := gin.New()

	router.GET("/manage/health", func (c *gin.Context) {
		c.Status(http.StatusOK)
	})

	api := router.Group("/api/v1") 
	{
		
		cars := api.Group("/cars")
		{
			cars.GET("", h.GetCars)
			cars.GET("/:uid", h.GetCarById)
			cars.POST("/query", h.GetCarsBatch)
			cars.PATCH("/:uid", h.UpdateCar)
		}
	}

	return router
}