package models

type Payment struct {
    ID        uint          `json:"id" gorm:"primaryKey;autoIncrement"`
    PaymentUID string       `json:"payment_uid" gorm:"type:uuid;not null"`
    Status     string       `json:"status" gorm:"type:varchar(20);not null;check:type IN ('PAID', 'CANCELED')"`
    Price      int          `json:"price" gorm:"type:integer;not null"`
}

func (Payment) TableName() string {
    return "payment"
}