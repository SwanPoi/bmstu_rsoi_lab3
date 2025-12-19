package models


type CarResponse struct {
    CarUID           string `json:"carUid"`
    Brand            string `json:"brand"`
    Model            string `json:"model"`
    RegistrationNumber string `json:"registrationNumber"`
    Power            int    `json:"power"`
    Type             string `json:"type"`
    Price            int    `json:"price"`
    Available        bool   `json:"available"`
}