package models

type RentCreation struct {
	PaymentUID	string 		`json:"paymentUid"`
	CarUID 		string 		`json:"carUid"`
	DateFrom	string 		`json:"dateFrom"`
	DateTo		string 		`json:"dateTo"`
	Username	string 		`json:"username"`
}