package services

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/SwanPoi/bmstu_rsoi_lab2/src/payment/models"
)

type MockPaymentRepository struct {
	mock.Mock
}

func (m *MockPaymentRepository) GetPaymentByUid(uid string) (*models.PaymentResponse, error) {
	args := m.Called(uid)
	if payment := args.Get(0); payment != nil {
		return payment.(*models.PaymentResponse), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPaymentRepository) GetPaymentsByUids(uids []string) ([]models.PaymentResponse, error) {
	args := m.Called(uids)
	return args.Get(0).([]models.PaymentResponse), args.Error(1)
}

func (m *MockPaymentRepository) UpdatePayment(payment models.PaymentUpsert, uid string) (*models.PaymentResponse, error) {
	args := m.Called(payment, uid)
	if response := args.Get(0); response != nil {
		return response.(*models.PaymentResponse), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPaymentRepository) CreatePayment(payment models.Payment) error {
	args := m.Called(payment)
	return args.Error(0)
}

// Тест: GetPaymentByUid успешно возвращает платеж
func TestPaymentService_GetPaymentByUid_Success(t *testing.T) {
	mockRepo := new(MockPaymentRepository)
	service := NewPaymentService(mockRepo)

	uid := "test-uid"
	expectedPayment := &models.PaymentResponse{
		PaymentUID: uid,
		Status:     "PAID",
		Price:      1000,
	}

	mockRepo.On("GetPaymentByUid", uid).Return(expectedPayment, nil)

	payment, err := service.GetPaymentByUid(uid)

	assert.Nil(t, err)
	assert.Equal(t, expectedPayment, payment)
	mockRepo.AssertExpectations(t)
}

// Тест: GetPaymentByUid возвращает ошибку из репозитория
func TestPaymentService_GetPaymentByUid_Error(t *testing.T) {
	mockRepo := new(MockPaymentRepository)
	service := NewPaymentService(mockRepo)

	uid := "test-uid"
	expectedError := errors.New("database error")

	mockRepo.On("GetPaymentByUid", uid).Return((*models.PaymentResponse)(nil), expectedError)

	_, err := service.GetPaymentByUid(uid)

	assert.True(t, errors.Is(err, expectedError))
	mockRepo.AssertExpectations(t)
}

// Тест: GetPaymentsByUids успешно возвращает список платежей
func TestPaymentService_GetPaymentsByUids_Success(t *testing.T) {
	mockRepo := new(MockPaymentRepository)
	service := NewPaymentService(mockRepo)

	uids := []string{"uid1", "uid2"}
	expectedPayments := []models.PaymentResponse{
		{PaymentUID: "uid1", Status: "PAID", Price: 1000},
		{PaymentUID: "uid2", Status: "CANCELED", Price: 0},
	}

	mockRepo.On("GetPaymentsByUids", uids).Return(expectedPayments, nil)

	payments, err := service.GetPaymentsByUids(uids)

	assert.Nil(t, err)
	assert.Equal(t, expectedPayments, payments)
	mockRepo.AssertExpectations(t)
}

// Тест: GetPaymentsByUids возвращает ошибку из репозитория
func TestPaymentService_GetPaymentsByUids_Error(t *testing.T) {
	mockRepo := new(MockPaymentRepository)
	service := NewPaymentService(mockRepo)

	uids := []string{"uid1", "uid2"}
	expectedError := errors.New("database error")

	mockRepo.On("GetPaymentsByUids", uids).Return([]models.PaymentResponse{}, expectedError)

	_, err := service.GetPaymentsByUids(uids)

	assert.True(t, errors.Is(err, expectedError))
	mockRepo.AssertExpectations(t)
}

// Тест: UpdatePayment возвращает ошибку InvalidStatus при невалидном статусе
func TestPaymentService_UpdatePayment_InvalidStatus(t *testing.T) {
	mockRepo := new(MockPaymentRepository)
	service := NewPaymentService(mockRepo)

	paymentUpsert := models.PaymentUpsert{
		Status: "INVALID_STATUS",
	}
	uid := "test-uid"

	_, err := service.UpdatePayment(paymentUpsert, uid)

	assert.True(t, errors.Is(err, models.InvalidStatus))
	mockRepo.AssertExpectations(t)
}

// Тест: UpdatePayment успешно обновляет платеж с валидным статусом
func TestPaymentService_UpdatePayment_Success(t *testing.T) {
	mockRepo := new(MockPaymentRepository)
	service := NewPaymentService(mockRepo)

	paymentUpsert := models.PaymentUpsert{
		Status: "CANCELED",
	}
	uid := "test-uid"
	expectedResponse := &models.PaymentResponse{
		PaymentUID: uid,
		Status:     "CANCELED",
		Price:      0,
	}

	mockRepo.On("UpdatePayment", paymentUpsert, uid).Return(expectedResponse, nil)

	response, err := service.UpdatePayment(paymentUpsert, uid)

	assert.Nil(t, err)
	assert.Equal(t, expectedResponse, response)
	mockRepo.AssertExpectations(t)
}

// Тест: CreatePayment успешно создаёт платеж с правильной стоимостью (2 дня)
func TestPaymentService_CreatePayment_Success_TwoDays(t *testing.T) {
	mockRepo := new(MockPaymentRepository)
	service := NewPaymentService(mockRepo)

	originalDayCost := models.DayCost
	models.DayCost = 1000
	defer func() { models.DayCost = originalDayCost }()

	dateFrom := time.Now().Truncate(24 * time.Hour)
	dateTo := dateFrom.Add(48 * time.Hour)

	paymentCreate := models.PaymentCreate{
		DateFrom: dateFrom,
		DateTo:   dateTo,
	}

	expectedPayment := models.Payment{
		Status: "PAID",
		Price:  2000,
	}

	mockRepo.On("CreatePayment", mock.MatchedBy(func(payment models.Payment) bool {
		return payment.Status == expectedPayment.Status &&
			payment.Price == expectedPayment.Price &&
			payment.PaymentUID != ""
	})).Return(nil)

	response, err := service.CreatePayment(paymentCreate)

	assert.Nil(t, err)
	assert.Equal(t, "PAID", response.Status)
	assert.Equal(t, 2000, response.Price)
	assert.NotEmpty(t, response.PaymentUID)
	mockRepo.AssertExpectations(t)
}

// Тест: CreatePayment возвращает ошибку из репозитория
func TestPaymentService_CreatePayment_RepoError(t *testing.T) {
	mockRepo := new(MockPaymentRepository)
	service := NewPaymentService(mockRepo)

	originalDayCost := models.DayCost
	models.DayCost = 1000
	defer func() { models.DayCost = originalDayCost }()

	dateFrom := time.Now().Truncate(24 * time.Hour)
	dateTo := dateFrom.Add(24 * time.Hour)

	paymentCreate := models.PaymentCreate{
		DateFrom: dateFrom,
		DateTo:   dateTo,
	}

	expectedError := errors.New("database error")
	mockRepo.On("CreatePayment", mock.Anything).Return(expectedError)

	_, err := service.CreatePayment(paymentCreate)

	assert.True(t, errors.Is(err, expectedError))
	mockRepo.AssertExpectations(t)
}