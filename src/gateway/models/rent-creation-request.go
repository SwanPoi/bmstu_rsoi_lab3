package models

type RentCreationRequest struct {
	CarUID 		string `json:"carUid"`
	DateFrom	string `json:"dateFrom"`
	DateTo		string `json:"dateTo"`
}