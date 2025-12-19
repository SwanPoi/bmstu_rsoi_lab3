package services

import (
	"github.com/SwanPoi/bmstu_rsoi_lab2/src/car/models"
	repo "github.com/SwanPoi/bmstu_rsoi_lab2/src/car/repositories"
)

type ICarService interface {
	GetCars(page int, size int, showAll bool) (*models.PaginationResponse, error)
	GetCarByUid(uuid string) (*models.ShortCar, error)
	GetCarsByUids([]string) ([]models.ShortCar, error)
	UpdateCar(models.CarUpsert, string) (*models.ShortCar, error)
}

type Services struct {
	ICarService
}

func NewServices(repo repo.ICarRepo) *Services {
	return &Services{
		ICarService: NewCarService(repo),
	}
}