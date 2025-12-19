package repositories

import (
	"github.com/SwanPoi/bmstu_rsoi_lab2/src/car/models"
	"gorm.io/gorm"
)

type ICarRepo interface {
	GetCars(int, int, bool) ([]models.Car, int, error)
	GetCarByUid(string) (*models.Car, error)
	GetCarsByUids([]string) ([]models.Car, error)
	UpdateCar(models.CarUpsert, string) (*models.Car, error)
}

type Repository struct {
	ICarRepo
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		ICarRepo: NewCarPostgres(db),
	}
}