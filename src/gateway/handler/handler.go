package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	cb "github.com/SwanPoi/bmstu_rsoi_lab2/src/gateway/circuitBreaker"
	"github.com/SwanPoi/bmstu_rsoi_lab2/src/gateway/models"
	services "github.com/SwanPoi/bmstu_rsoi_lab2/src/gateway/services"
)

type GatewayHandler struct {
	services *services.Services
	config   *models.HandlerConfig
	carCB       *cb.CircuitBreaker
	rentalCB    *cb.CircuitBreaker
	paymentCB   *cb.CircuitBreaker
}

func NewHandler(services *services.Services, config *models.HandlerConfig) *GatewayHandler {
	return &GatewayHandler{
		services: services,
		config: config,
		carCB:     cb.NewCircuitBreaker(5, 0.4, 30*time.Second),
		rentalCB:  cb.NewCircuitBreaker(5, 0.4, 30*time.Second),
		paymentCB: cb.NewCircuitBreaker(5, 0.4, 30*time.Second),
	}
}

func (h *GatewayHandler) SetupRoutes() *gin.Engine {
	router := gin.New()

	router.GET("/manage/health", func (c *gin.Context) {
		c.Status(http.StatusOK)
	})

	api := router.Group("/api/v1") 
	{
		cars := api.Group("/cars") 
		{
			cars.GET("", h.GetCars)
		}

		rental := api.Group("/rental")
		{
			rental.GET("", h.GetUserRentals)
			rental.GET(":rentalUid", h.GetRentalById)

			rental.POST("", h.RentCar)
			rental.POST(":rentalUid/finish", h.FinishCarRent)

			rental.DELETE(":rentalUid", h.RevokeRent)
		}
	}

	return router
}