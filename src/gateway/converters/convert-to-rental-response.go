package converters

import "github.com/SwanPoi/bmstu_rsoi_lab2/src/gateway/models"

func ConvertToRentalResponse(rental models.RentalInfo, car models.CarInfo, payment models.PaymentInfo) models.RentalResponse {
	response := models.RentalResponse{
		RentalUID: rental.RentalUID,
		Status: rental.Status,
		DateFrom: rental.DateFrom,
		DateTo: rental.DateTo,
		Car: car,
		Payment: payment,
	}

	return response
}