package utils

import "github.com/SwanPoi/bmstu_rsoi_lab2/src/rental/models"

func ConvertToRentalResponse(rental models.Rental) models.RentalResponse {
	return models.RentalResponse{
		RentalUID: rental.RentalUID,
		PaymentUID: rental.PaymentUID,
		CarUID: rental.CarUID,
		DateFrom: rental.DateFrom.Format("2006-01-02"),
		DateTo: rental.DateTo.Format("2006-01-02"),
		Status: rental.Status,
	}
}