package services

import (
	"github.com/SwanPoi/bmstu_rsoi_lab2/src/rental/models"
	repo "github.com/SwanPoi/bmstu_rsoi_lab2/src/rental/repositories"
)

type IRentalService interface {
	GetUserRentalByUid(uid string, username string) (*models.RentalResponse, error)
	GetUserRentals(username string) ([]models.RentalResponse, error)
	CreateRental(models.RentCreation) (*models.RentalResponse, error)
	UpdateRental(rental models.RentalUpsert, uid string, username string) (*models.RentalResponse, error)
}

type Services struct {
	IRentalService
}

func NewServices(repo *repo.Repository) *Services {
	return &Services{
		IRentalService: NewRentalService(repo),
	}
}