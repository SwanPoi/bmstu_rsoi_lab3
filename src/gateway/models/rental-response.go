package models

type RentalResponse struct {
	RentalUID string    		`json:"rentalUid"`
    DateFrom  string 			`json:"dateFrom"`
    DateTo    string 			`json:"dateTo"`
    Status    string    		`json:"status"`
	Car		  CarInfo 			`json:"car"`
	Payment   PaymentInfo		`json:"payment"`
}