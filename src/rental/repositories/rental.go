package repositories

import (
	"errors"

	"github.com/SwanPoi/bmstu_rsoi_lab2/src/rental/models"
	"github.com/SwanPoi/bmstu_rsoi_lab2/src/rental/utils"
	"gorm.io/gorm"
)

type RentalPostgres struct {
	DB *gorm.DB
}

func NewRentalPostgres(db *gorm.DB) *RentalPostgres {
	return &RentalPostgres{DB: db}
}

func (r *RentalPostgres) GetRentalByUid(uid string) (*models.Rental, error) {
	var rental models.Rental

	if err := r.DB.Where("rental_uid = ?", uid).First(&rental).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, models.ErrorNotFound
		}

		return nil, err
	}

	return &rental, nil
}

func (r *RentalPostgres) GetUserRentals(username string) ([]models.RentalResponse, error) {
	var rentals []models.Rental

	if err := r.DB.Omit("id", "username").Where("username = ?", username).Find(&rentals).Error; err != nil {
		return nil, err
	}

	responses := make([]models.RentalResponse, len(rentals))

	for i, rental := range rentals {
		responses[i] = utils.ConvertToRentalResponse(rental)
	}

	return responses, nil
}

func (r *RentalPostgres) CreateRental(rental models.Rental) (error) {
	return r.DB.Create(&rental).Error
}

func (r *RentalPostgres) UpdateRental(rentalUpsert models.RentalUpsert, uid string, username string) (*models.RentalResponse, error) {
	result := r.DB.Model(&models.Rental{}).
					Where("rental_uid = ? AND username = ?", uid, username).
					Update("status", rentalUpsert.Status)
	
	if result.Error != nil {
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		return nil, models.ErrorNotFound
	}

	var rental models.Rental

	if err := r.DB.Omit("id", "username").Where("rental_uid = ? AND username = ?", uid, username).Find(&rental).Error; err != nil {
		return nil, err
	}

	updatedRental := utils.ConvertToRentalResponse(rental)

	return &updatedRental, nil
}