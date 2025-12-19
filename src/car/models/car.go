package models

type Car struct {
    ID                uint      `json:"id" gorm:"primaryKey;autoIncrement"`
    CarUID            string    `json:"car_uid" gorm:"type:uuid;uniqueIndex;not null"`
    Brand             string    `json:"brand" gorm:"type:varchar(80);not null"`
    Model             string    `json:"model" gorm:"type:varchar(80);not null"`
    RegistrationNumber string   `json:"registration_number" gorm:"type:varchar(20);not null"`
    Power             int       `json:"power" gorm:"type:integer"`
    Price             int       `json:"price" gorm:"type:integer;not null"`
    Type              string    `json:"type" gorm:"type:varchar(20);check:type IN ('SEDAN', 'SUV', 'MINIVAN', 'ROADSTER')"`
    Availability      bool      `json:"availability" gorm:"not null"`
}