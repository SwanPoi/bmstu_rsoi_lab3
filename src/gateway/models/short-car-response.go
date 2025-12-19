package models

type ShortCarResponse struct {
	CarUID              string `json:"carUid"`
    Brand               string `json:"brand"`
    Model               string `json:"model"`
    RegistrationNumber  string `json:"registrationNumber"`
    Availability        bool   `json:"availability"`
}