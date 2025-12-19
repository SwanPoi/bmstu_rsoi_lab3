package models

import "time"

type PaymentCreate struct {
	DateFrom	time.Time `json:"dateFrom"`
	DateTo		time.Time `json:"dateTo"`
}