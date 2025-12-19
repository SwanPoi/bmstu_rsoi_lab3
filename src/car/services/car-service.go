package services

import (
	"github.com/SwanPoi/bmstu_rsoi_lab2/src/car/converters"
	"github.com/SwanPoi/bmstu_rsoi_lab2/src/car/models"
	repo "github.com/SwanPoi/bmstu_rsoi_lab2/src/car/repositories"
)

type CarService struct {
	repo repo.ICarRepo
}

func NewCarService(repo repo.ICarRepo) *CarService {
	return &CarService{repo: repo}
}

func (s *CarService) GetCars(page int, size int, showAll bool) (*models.PaginationResponse, error) {
	offset := (page - 1) * size

	cars, total, err := s.repo.GetCars(offset, size, showAll)

	if err != nil {
		return nil, err
	}

	carsResponse := converters.CarResponsesFromCars(cars)

	paginationResponse := &models.PaginationResponse{
		Page: page,
		PageSize: size,
		TotalElements: total,
		Items: carsResponse,
	}

	return paginationResponse, nil
}

func (s *CarService) GetCarByUid(uid string) (*models.ShortCar, error) {
	car, err := s.repo.GetCarByUid(uid)

	if err != nil {
		return nil, err
	}

	shortCar := converters.CarToShortCar(*car)

	return &shortCar, nil
}

func (s *CarService) GetCarsByUids(uids []string) ([]models.ShortCar, error) {
	cars, err := s.repo.GetCarsByUids(uids)

	if err != nil {
		return nil, err
	}

	var shortCars = make([]models.ShortCar, len(cars))

	for i, car := range cars {
		shortCars[i] = converters.CarToShortCar(car)
	}

	return shortCars, nil
}

func (s *CarService) UpdateCar(car models.CarUpsert, uid string) (*models.ShortCar, error) {
	updatedCar, err := s.repo.UpdateCar(car, uid); 
	if err != nil {
		return nil, err
	}

	shortCar := converters.CarToShortCar(*updatedCar)

	return &shortCar, nil
}