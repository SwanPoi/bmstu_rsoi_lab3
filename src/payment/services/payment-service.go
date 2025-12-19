package services

import (
	"math"
	"time"

	"github.com/SwanPoi/bmstu_rsoi_lab2/src/payment/models"
	repo "github.com/SwanPoi/bmstu_rsoi_lab2/src/payment/repositories"
	"github.com/google/uuid"
)

type PaymentService struct {
	repo repo.IPaymentRepo
}

func NewPaymentService(repo repo.IPaymentRepo) *PaymentService {
	return &PaymentService{repo: repo}
}

func (s *PaymentService) GetPaymentByUid(uid string) (*models.PaymentResponse, error) {
	return s.repo.GetPaymentByUid(uid)
}

func (s *PaymentService) GetPaymentsByUids(uids []string) ([]models.PaymentResponse, error) {
	return s.repo.GetPaymentsByUids(uids)
}

func (s *PaymentService) UpdatePayment(payment models.PaymentUpsert, uid string) (*models.PaymentResponse, error) {
	validStatuses := map[string]bool{
        "PAID": true,
        "CANCELED":    true,
    }

	if !validStatuses[payment.Status] {
        return nil, models.InvalidStatus
    }

	return s.repo.UpdatePayment(payment, uid)
}

func (s *PaymentService) CreatePayment(paymentInsert models.PaymentCreate) (*models.PaymentResponse, error) {
	duration := paymentInsert.DateTo.Sub(paymentInsert.DateFrom)
	days := int(math.Round(duration.Round(time.Hour).Hours() / 24))

	payment := models.Payment{
		PaymentUID: uuid.New().String(),
		Status: "PAID",
		Price: models.DayCost * days,
	}

	if err := s.repo.CreatePayment(payment); err == nil {
		response := models.PaymentResponse{
			PaymentUID: payment.PaymentUID,
			Status: payment.Status,
			Price: payment.Price,
		}

		return &response, nil
	} else {
		return nil, err
	}
}
