package models

type CreateRentalResponse struct {
	RentalUID string    				`json:"rentalUid"`
    CarUID    string    				`json:"carUid"`
    DateFrom  string 					`json:"dateFrom"`
    DateTo    string 					`json:"dateTo"`
    Status    string    				`json:"status"`
	Payment   PaymentCreationResponse	`json:"payment"`
}