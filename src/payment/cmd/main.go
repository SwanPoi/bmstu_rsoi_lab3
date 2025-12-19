package main

import (
	"log"

	handler "github.com/SwanPoi/bmstu_rsoi_lab2/src/payment/handler"
	"github.com/SwanPoi/bmstu_rsoi_lab2/src/payment/models"
	repo "github.com/SwanPoi/bmstu_rsoi_lab2/src/payment/repositories"
	server "github.com/SwanPoi/bmstu_rsoi_lab2/src/payment/server"
	services "github.com/SwanPoi/bmstu_rsoi_lab2/src/payment/services"
)

func main() {
	connString := repo.GetConnectionString(&repo.DatabaseConfig{
		Host: "postgres",
		Port: 5432,
		User: "postgres",
		Password: "postgres",
		Database: "payments",
	})

	db, err := repo.InitDb(connString)

	if err != nil {
		log.Fatal("Fail during db connection", err)
		return
	}

	log.Print("Successfully connect to database")
	db.AutoMigrate(&models.Payment{})

	repos := repo.NewRepository(db)
	service := services.NewServices(repos)
	handler := handler.NewHandler(service)

	srv := new(server.CommonServer)

	if err := srv.Run("8050", handler.SetupRoutes()); err != nil {
		log.Fatal("Fail during payment server start", err)
		return
	}
}