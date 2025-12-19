package services

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/SwanPoi/bmstu_rsoi_lab2/src/rental/models"
	"github.com/SwanPoi/bmstu_rsoi_lab2/src/rental/utils"
)

type MockRentalRepository struct {
	mock.Mock
}

func (m *MockRentalRepository) GetRentalByUid(uid string) (*models.Rental, error) {
	args := m.Called(uid)
	if rental := args.Get(0); rental != nil {
		return rental.(*models.Rental), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockRentalRepository) GetUserRentals(username string) ([]models.RentalResponse, error) {
	args := m.Called(username)
	return args.Get(0).([]models.RentalResponse), args.Error(1)
}

func (m *MockRentalRepository) CreateRental(rental models.Rental) error {
	args := m.Called(rental)
	return args.Error(0)
}

func (m *MockRentalRepository) UpdateRental(rental models.RentalUpsert, uid string, username string) (*models.RentalResponse, error) {
	args := m.Called(rental, uid, username)
	if response := args.Get(0); response != nil {
		return response.(*models.RentalResponse), args.Error(1)
	}
	return nil, args.Error(1)
}

// Тест: GetUserRentalByUid возвращает ошибку, если запись не найдена
func TestRentalService_GetUserRentalByUid_NotFound(t *testing.T) {
	mockRepo := new(MockRentalRepository)
	service := NewRentalService(mockRepo)

	uid := "test-uid"
	username := "test-user"
	expectedError := errors.New("record not found")

	mockRepo.On("GetRentalByUid", uid).Return((*models.Rental)(nil), expectedError)

	_, err := service.GetUserRentalByUid(uid, username)

	assert.True(t, errors.Is(err, expectedError))
	mockRepo.AssertExpectations(t)
}

// Тест: GetUserRentalByUid возвращает ошибку Forbidden, если username не совпадает
func TestRentalService_GetUserRentalByUid_Forbidden(t *testing.T) {
	mockRepo := new(MockRentalRepository)
	service := NewRentalService(mockRepo)

	uid := "test-uid"
	username := "john_doe"
	rental := &models.Rental{
		RentalUID: uid,
		Username:  "jane_smith",
	}

	mockRepo.On("GetRentalByUid", uid).Return(rental, nil)

	_, err := service.GetUserRentalByUid(uid, username)

	assert.True(t, errors.Is(err, models.Forbidden))
	mockRepo.AssertExpectations(t)
}

// Тест: GetUserRentalByUid успешно возвращает запись
func TestRentalService_GetUserRentalByUid_Success(t *testing.T) {
	mockRepo := new(MockRentalRepository)
	service := NewRentalService(mockRepo)

	uid := "test-uid"
	username := "john_doe"
	rental := &models.Rental{
		RentalUID: uid,
		Username:  username,
		Status:    "IN_PROGRESS",
	}

	expectedResponse := utils.ConvertToRentalResponse(*rental)

	mockRepo.On("GetRentalByUid", uid).Return(rental, nil)

	response, err := service.GetUserRentalByUid(uid, username)

	assert.Nil(t, err)
	assert.Equal(t, expectedResponse, *response)
	mockRepo.AssertExpectations(t)
}

// Тест: GetUserRentals успешно возвращает список аренд
func TestRentalService_GetUserRentals_Success(t *testing.T) {
	mockRepo := new(MockRentalRepository)
	service := NewRentalService(mockRepo)

	username := "john_doe"
	expectedRentals := []models.RentalResponse{
		{RentalUID: "uid1", Status: "IN_PROGRESS"},
		{RentalUID: "uid2", Status: "FINISHED"},
	}

	mockRepo.On("GetUserRentals", username).Return(expectedRentals, nil)

	rentals, err := service.GetUserRentals(username)

	assert.Nil(t, err)
	assert.Equal(t, expectedRentals, rentals)
	mockRepo.AssertExpectations(t)
}

// Тест: GetUserRentals возвращает ошибку из репозитория
func TestRentalService_GetUserRentals_Error(t *testing.T) {
	mockRepo := new(MockRentalRepository)
	service := NewRentalService(mockRepo)

	username := "john_doe"
	expectedError := errors.New("database error")

	mockRepo.On("GetUserRentals", username).Return([]models.RentalResponse{}, expectedError)

	_, err := service.GetUserRentals(username)

	assert.True(t, errors.Is(err, expectedError))
	mockRepo.AssertExpectations(t)
}

// Тест: CreateRental успешно создаёт запись
func TestRentalService_CreateRental_Success(t *testing.T) {
	mockRepo := new(MockRentalRepository)
	service := NewRentalService(mockRepo)

	rentalReq := models.RentCreation{
		Username:    "john_doe",
		CarUID:      "car-uid",
		PaymentUID:  "payment-uid",
		DateFrom:    "2023-12-01",
		DateTo:      "2023-12-05",
	}

	expectedRental := models.Rental{
		Username:   rentalReq.Username,
		CarUID:     rentalReq.CarUID,
		PaymentUID: rentalReq.PaymentUID,
		Status:     "IN_PROGRESS",
	}
	expectedRental.RentalUID = uuid.New().String()
	dateFrom, _ := time.Parse("2006-01-02", rentalReq.DateFrom)
	dateTo, _ := time.Parse("2006-01-02", rentalReq.DateTo)
	expectedRental.DateFrom = dateFrom
	expectedRental.DateTo = dateTo

	expectedResponse := utils.ConvertToRentalResponse(expectedRental)

	mockRepo.On("CreateRental", mock.MatchedBy(func(rental models.Rental) bool {
		return rental.Username == expectedRental.Username &&
			rental.CarUID == expectedRental.CarUID &&
			rental.PaymentUID == expectedRental.PaymentUID &&
			rental.Status == expectedRental.Status &&
			rental.DateFrom.Equal(expectedRental.DateFrom) &&
			rental.DateTo.Equal(expectedRental.DateTo)
	})).Return(nil)

	response, err := service.CreateRental(rentalReq)

	assert.Nil(t, err)
	assert.Equal(t, expectedResponse.CarUID, response.CarUID)
	mockRepo.AssertExpectations(t)
}

// Тест: CreateRental возвращает ошибку при невалидной дате
func TestRentalService_CreateRental_InvalidDate(t *testing.T) {
	mockRepo := new(MockRentalRepository)
	service := NewRentalService(mockRepo)

	rentalReq := models.RentCreation{
		Username: "john_doe",
		CarUID:   "car-uid",
		DateFrom: "invalid-date",
		DateTo:   "2023-12-05",
	}

	_, err := service.CreateRental(rentalReq)

	assert.NotNil(t, err)
	mockRepo.AssertExpectations(t)
}
