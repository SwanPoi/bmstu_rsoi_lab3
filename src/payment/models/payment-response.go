package models

type PaymentResponse struct {
    PaymentUID string       `json:"paymentUid"`
    Status     string       `json:"status"`
    Price      int          `json:"price"`
}

func (PaymentResponse) TableName() string {
    return "payment"
}