package models

type PaymentInfo struct {
	PaymentUID string       `json:"paymentUid"`
    Status     string       `json:"status"`
    Price      int          `json:"price"`
}