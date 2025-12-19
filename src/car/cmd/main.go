package main

import (
	"log"

	handler "github.com/SwanPoi/bmstu_rsoi_lab2/src/car/handler"
	"github.com/SwanPoi/bmstu_rsoi_lab2/src/car/models"
	repo "github.com/SwanPoi/bmstu_rsoi_lab2/src/car/repositories"
	server "github.com/SwanPoi/bmstu_rsoi_lab2/src/car/server"
	services "github.com/SwanPoi/bmstu_rsoi_lab2/src/car/services"
)

func main() {
	connString := repo.GetConnectionString(&repo.DatabaseConfig{
		Host: "postgres",
		Port: 5432,
		User: "postgres",
		Password: "postgres",
		Database: "cars",
	})

	db, err := repo.InitDb(connString)

	if err != nil {
		log.Fatal("Fail during db connection", err)
		return
	}

	log.Print("Successfully connect to database")
	db.AutoMigrate(&models.Car{})

	repos := repo.NewRepository(db)
	service := services.NewServices(repos)
	handler := handler.NewHandler(service)

	srv := new(server.CommonServer)

	if err := srv.Run("8070", handler.SetupRoutes()); err != nil {
		log.Fatal("Fail during car server start", err)
		return
	}
}