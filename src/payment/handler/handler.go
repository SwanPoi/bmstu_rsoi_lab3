package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	services "github.com/SwanPoi/bmstu_rsoi_lab2/src/payment/services"
)

type PaymentHandler struct {
	services *services.Services
}

func NewHandler(services *services.Services) *PaymentHandler {
	return &PaymentHandler{services: services}
}

func (h *PaymentHandler) SetupRoutes() *gin.Engine {
	router := gin.New()

	router.GET("/manage/health", func (c *gin.Context) {
		c.Status(http.StatusOK)
	})

	api := router.Group("/api/v1") 
	{
		payments := api.Group("/payment")
		{
			payments.POST("", h.CreatePayment)
			payments.GET("/:uid", h.GetPaymentByUid)
			payments.POST("/query", h.GetPaymensBatch)
			payments.PATCH("/:uid", h.UpdatePayment)
		}
	}

	return router
}