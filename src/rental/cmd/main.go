package main

import (
	"log"

	handler "github.com/SwanPoi/bmstu_rsoi_lab2/src/rental/handler"
	"github.com/SwanPoi/bmstu_rsoi_lab2/src/rental/models"
	repo "github.com/SwanPoi/bmstu_rsoi_lab2/src/rental/repositories"
	server "github.com/SwanPoi/bmstu_rsoi_lab2/src/rental/server"
	services "github.com/SwanPoi/bmstu_rsoi_lab2/src/rental/services"
)

func main() {
	connString := repo.GetConnectionString(&repo.DatabaseConfig{
		Host: "postgres",
		Port: 5432,
		User: "postgres",
		Password: "postgres",
		Database: "rentals",
	})

	db, err := repo.InitDb(connString)

	if err != nil {
		log.Fatal("Fail during db connection", err)
		return
	}

	log.Print("Successfully connect to database")
	db.AutoMigrate(&models.Rental{})

	repos := repo.NewRepository(db)
	service := services.NewServices(repos)
	handler := handler.NewHandler(service)

	srv := new(server.CommonServer)

	if err := srv.Run("8060", handler.SetupRoutes()); err != nil {
		log.Fatal("Fail during rental server start", err)
		return
	}
}