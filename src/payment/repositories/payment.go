package repositories

import (
	"errors"

	"github.com/SwanPoi/bmstu_rsoi_lab2/src/payment/models"
	"gorm.io/gorm"
)

type PaymentPostgres struct {
	DB *gorm.DB
}

func NewPaymentPostgres(db *gorm.DB) *PaymentPostgres {
	return &PaymentPostgres{DB: db}
}

func (r *PaymentPostgres) GetPaymentByUid(uid string) (*models.PaymentResponse, error) {
	var payment models.PaymentResponse

	if err := r.DB.Select("payment_uid", "status", "price").Where("payment_uid = ?", uid).First(&payment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, models.ErrorNotFound
		}

		return nil, err
	}

	return &payment, nil
}

func (r *PaymentPostgres) GetPaymentsByUids(uids []string) ([]models.PaymentResponse, error) {
	var payments []models.PaymentResponse

	if err := r.DB.Select("payment_uid", "status", "price").Where("payment_uid IN ?", uids).Find(&payments).Error; err != nil {
		return nil, err
	}

	return payments, nil
}

func (r *PaymentPostgres) UpdatePayment(payment models.PaymentUpsert, uid string) (*models.PaymentResponse, error) {
	result := r.DB.Model(&models.Payment{}).
				Where("payment_uid = ?", uid).
				Update("status", payment.Status)
	
	if result.Error != nil {
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		return nil, models.ErrorNotFound
	}

	var updatedPayment models.PaymentResponse

	if err := r.DB.Select("payment_uid", "status", "price").Where("payment_uid = ?", uid).First(&updatedPayment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, models.ErrorNotFound
		}

		return nil, err
	}

	return &updatedPayment, nil
}

func (r *PaymentPostgres) CreatePayment(payment models.Payment) (error) {
	return r.DB.Create(&payment).Error
}
