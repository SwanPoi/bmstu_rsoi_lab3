package models

type PaymentCreateRequest struct {
	DateFrom	string `json:"dateFrom"`
	DateTo		string `json:"dateTo"`
}