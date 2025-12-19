package services

import (
	"time"

	"github.com/SwanPoi/bmstu_rsoi_lab2/src/rental/models"
	repo "github.com/SwanPoi/bmstu_rsoi_lab2/src/rental/repositories"
	"github.com/SwanPoi/bmstu_rsoi_lab2/src/rental/utils"
	"github.com/google/uuid"
)

type RentalService struct {
	repo repo.IRentalRepo
}

func NewRentalService(repo repo.IRentalRepo) *RentalService {
	return &RentalService{repo: repo}
}

func (s *RentalService) GetUserRentalByUid(uid string, username string) (*models.RentalResponse, error) {
	rental, err := s.repo.GetRentalByUid(uid)

	if err != nil {
		return nil, err
	}

	if rental.Username != username {
		return nil, models.Forbidden
	}

	rentalResponse := utils.ConvertToRentalResponse(*rental)

	return &rentalResponse, nil
}

func (s *RentalService) GetUserRentals(username string) ([]models.RentalResponse, error) {
	return s.repo.GetUserRentals(username)
}

func (s *RentalService) CreateRental(rentalReq models.RentCreation) (*models.RentalResponse, error) {
	dateFrom, err := time.Parse("2006-01-02", rentalReq.DateFrom)
    if err != nil {
        return nil, err
    }

    dateTo, err := time.Parse("2006-01-02", rentalReq.DateTo)
    if err != nil {
        return nil, err
    }

	rental := models.Rental{
		RentalUID: uuid.New().String(),
		Username: rentalReq.Username,
		CarUID: rentalReq.CarUID,
		PaymentUID: rentalReq.PaymentUID,
		Status: "IN_PROGRESS",
		DateFrom: dateFrom,
		DateTo: dateTo,
	}

	if err := s.repo.CreateRental(rental); err == nil {
		response := utils.ConvertToRentalResponse(rental)
		return &response, nil
	} else {
		return nil, err
	}
}

func (s *RentalService) UpdateRental(rental models.RentalUpsert, uid string, username string) (*models.RentalResponse, error) {
	validStatuses := map[string]bool{
        "IN_PROGRESS": true,
        "FINISHED":    true,
        "CANCELED":    true,
    }

	if !validStatuses[rental.Status] {
        return nil, models.InvalidStatus
    }

	return s.repo.UpdateRental(rental, uid, username)
}