package converters

import "github.com/SwanPoi/bmstu_rsoi_lab2/src/gateway/models"

func ConvertToCreateRentalResponse(rental models.RentalInfo, payment models.PaymentCreationResponse) models.CreateRentalResponse {
	return models.CreateRentalResponse{
		RentalUID: rental.RentalUID,
		Status: rental.Status,
		DateFrom: rental.DateFrom,
		DateTo: rental.DateTo,
		CarUID: rental.CarUID,
		Payment: payment,
	}
}