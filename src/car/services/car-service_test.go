package services

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/SwanPoi/bmstu_rsoi_lab2/src/car/converters"
	"github.com/SwanPoi/bmstu_rsoi_lab2/src/car/models"
)

type MockCarRepository struct {
	mock.Mock
}

func (m *MockCarRepository) GetCars(offset int, size int, showAll bool) ([]models.Car, int, error) {
    args := m.Called(offset, size, showAll)
    return args.Get(0).([]models.Car), args.Get(1).(int), args.Error(2)
}

func (m *MockCarRepository) GetCarByUid(uid string) (*models.Car, error) {
	args := m.Called(uid)
	if car := args.Get(0); car != nil {
		return car.(*models.Car), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockCarRepository) GetCarsByUids(uids []string) ([]models.Car, error) {
	args := m.Called(uids)
	return args.Get(0).([]models.Car), args.Error(1)
}

func (m *MockCarRepository) UpdateCar(car models.CarUpsert, uid string) (*models.Car, error) {
	args := m.Called(car, uid)
	if updatedCar := args.Get(0); updatedCar != nil {
		return updatedCar.(*models.Car), args.Error(1)
	}
	return nil, args.Error(1)
}

// Тест: GetCars успешно возвращает пагинированный список автомобилей (showAll = false)
func TestCarService_GetCars_Success_ShowAllFalse(t *testing.T) {
	mockRepo := new(MockCarRepository)
	service := NewCarService(mockRepo)

	page := 1
	size := 10
	showAll := false
	offset := (page - 1) * size

	cars := []models.Car{
		{CarUID: "uid1", Brand: "Toyota", Model: "Camry", Availability: true},
		{CarUID: "uid2", Brand: "BMW", Model: "X5", Availability: true},
	}
	total := 2
	carsResponse := converters.CarResponsesFromCars(cars)

	mockRepo.On("GetCars", offset, size, showAll).Return(cars, total, nil)

	result, err := service.GetCars(page, size, showAll)

	assert.Nil(t, err)
	assert.Equal(t, page, result.Page)
	assert.Equal(t, size, result.PageSize)
	assert.Equal(t, total, result.TotalElements)
	assert.Equal(t, carsResponse, result.Items)
	mockRepo.AssertExpectations(t)
}

// Тест: GetCars успешно возвращает пагинированный список автомобилей (showAll = true)
func TestCarService_GetCars_Success_ShowAllTrue(t *testing.T) {
	mockRepo := new(MockCarRepository)
	service := NewCarService(mockRepo)

	page := 2
	size := 5
	showAll := true
	offset := (page - 1) * size

	cars := []models.Car{
		{CarUID: "uid3", Brand: "Honda", Model: "Civic", Availability: false},
		{CarUID: "uid4", Brand: "Ford", Model: "Focus", Availability: true},
	}
	total := 15
	carsResponse := converters.CarResponsesFromCars(cars)

	mockRepo.On("GetCars", offset, size, showAll).Return(cars, total, nil)

	result, err := service.GetCars(page, size, showAll)

	assert.Nil(t, err)
	assert.Equal(t, page, result.Page)
	assert.Equal(t, size, result.PageSize)
	assert.Equal(t, total, result.TotalElements)
	assert.Equal(t, carsResponse, result.Items)
	mockRepo.AssertExpectations(t)
}

// Тест: GetCarByUid успешно возвращает автомобиль
func TestCarService_GetCarByUid_Success(t *testing.T) {
	mockRepo := new(MockCarRepository)
	service := NewCarService(mockRepo)

	uid := "test-uid"
	car := &models.Car{
		CarUID: uid,
		Brand:  "Toyota",
		Model:  "Camry",
		Availability: true,
	}
	expectedShortCar := converters.CarToShortCar(*car)

	mockRepo.On("GetCarByUid", uid).Return(car, nil)

	result, err := service.GetCarByUid(uid)

	assert.Nil(t, err)
	assert.Equal(t, expectedShortCar, *result)
	mockRepo.AssertExpectations(t)
}

// Тест: GetCarByUid возвращает ошибку из репозитория
func TestCarService_GetCarByUid_RepoError(t *testing.T) {
	mockRepo := new(MockCarRepository)
	service := NewCarService(mockRepo)

	uid := "test-uid"
	expectedError := errors.New("car not found")

	mockRepo.On("GetCarByUid", uid).Return((*models.Car)(nil), expectedError)

	_, err := service.GetCarByUid(uid)

	assert.True(t, errors.Is(err, expectedError))
	mockRepo.AssertExpectations(t)
}

// Тест: GetCarsByUids успешно возвращает список автомобилей
func TestCarService_GetCarsByUids_Success(t *testing.T) {
	mockRepo := new(MockCarRepository)
	service := NewCarService(mockRepo)

	uids := []string{"uid1", "uid2", "uid3"}
	cars := []models.Car{
		{CarUID: "uid1", Brand: "Toyota", Model: "Camry"},
		{CarUID: "uid2", Brand: "BMW", Model: "X5"},
		{CarUID: "uid3", Brand: "Honda", Model: "Civic"},
	}

	expectedShortCars := make([]models.ShortCar, len(cars))
	for i, car := range cars {
		expectedShortCars[i] = converters.CarToShortCar(car)
	}

	mockRepo.On("GetCarsByUids", uids).Return(cars, nil)

	result, err := service.GetCarsByUids(uids)

	assert.Nil(t, err)
	assert.Equal(t, expectedShortCars, result)
	mockRepo.AssertExpectations(t)
}

// Тест: GetCarsByUids возвращает ошибку из репозитория
func TestCarService_GetCarsByUids_RepoError(t *testing.T) {
	mockRepo := new(MockCarRepository)
	service := NewCarService(mockRepo)

	uids := []string{"uid1", "uid2"}
	expectedError := errors.New("database error")

	mockRepo.On("GetCarsByUids", uids).Return([]models.Car{}, expectedError)

	_, err := service.GetCarsByUids(uids)

	assert.True(t, errors.Is(err, expectedError))
	mockRepo.AssertExpectations(t)
}

// Тест: UpdateCar успешно обновляет автомобиль
func TestCarService_UpdateCar_Success(t *testing.T) {
	mockRepo := new(MockCarRepository)
	service := NewCarService(mockRepo)

	uid := "test-uid"
	carUpsert := models.CarUpsert{
		Availability: false,
	}
	updatedCar := &models.Car{
		CarUID:       uid,
		Brand:        "Toyota",
		Model:        "Camry",
		Availability: false,
	}
	expectedShortCar := converters.CarToShortCar(*updatedCar)

	mockRepo.On("UpdateCar", carUpsert, uid).Return(updatedCar, nil)

	result, err := service.UpdateCar(carUpsert, uid)

	assert.Nil(t, err)
	assert.Equal(t, expectedShortCar, *result)
	mockRepo.AssertExpectations(t)
}

// Тест: UpdateCar возвращает ошибку из репозитория
func TestCarService_UpdateCar_RepoError(t *testing.T) {
	mockRepo := new(MockCarRepository)
	service := NewCarService(mockRepo)

	uid := "test-uid"
	carUpsert := models.CarUpsert{
		Availability: false,
	}
	expectedError := errors.New("update failed")

	mockRepo.On("UpdateCar", carUpsert, uid).Return((*models.Car)(nil), expectedError)

	_, err := service.UpdateCar(carUpsert, uid)

	assert.True(t, errors.Is(err, expectedError))
	mockRepo.AssertExpectations(t)
}