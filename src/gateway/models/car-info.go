package models

type CarInfo struct {
	CarUID              string `json:"carUid"`
    Brand               string `json:"brand"`
    Model               string `json:"model"`
    RegistrationNumber  string `json:"registrationNumber"`
}