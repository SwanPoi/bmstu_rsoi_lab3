package models

import "time"

type Rental struct {
    ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
    RentalUID string    `json:"rental_uid" gorm:"type:uuid;uniqueIndex;not null"`
    Username  string    `json:"username" gorm:"type:varchar(80);not null"`
    PaymentUID string   `json:"payment_uid" gorm:"type:uuid;not null"`
    CarUID    string    `json:"car_uid" gorm:"type:uuid;not null"`
    DateFrom  time.Time `json:"date_from" gorm:"type:timestamp with time zone;not null"`
    DateTo    time.Time `json:"date_to" gorm:"type:timestamp with time zone;not null"`
    Status    string    `json:"status" gorm:"type:varchar(20);not null;check:type IN ('IN_PROGRESS', 'FINISHED', 'CANCELED')"`
}

func (Rental) TableName() string {
    return "rental"
}