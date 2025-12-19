package main

import (
	"log"

	handler "github.com/SwanPoi/bmstu_rsoi_lab2/src/gateway/handler"
	"github.com/SwanPoi/bmstu_rsoi_lab2/src/gateway/models"
	server "github.com/SwanPoi/bmstu_rsoi_lab2/src/gateway/server"
	services "github.com/SwanPoi/bmstu_rsoi_lab2/src/gateway/services"
	redis 	"github.com/SwanPoi/bmstu_rsoi_lab2/src/gateway/queue"
)

func main() {
	redis.InitRedis()
	redis.StartRetryWorker()
	
	services := services.NewServices()

	handlerConfig := models.HandlerConfig{
		CarUrl: "http://cars:8070/api/v1",
		RentalUrl: "http://rental:8060/api/v1",
		PaymentUrl: "http://payment:8050/api/v1",
	}

	handler := handler.NewHandler(services, &handlerConfig)

	srv := new(server.CommonServer)


	if err := srv.Run("8080", handler.SetupRoutes()); err != nil {
		log.Fatal("Fail during gateway server start: ", err)
		return
	}
}