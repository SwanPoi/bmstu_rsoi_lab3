package converters

import "github.com/SwanPoi/bmstu_rsoi_lab2/src/car/models"

func CarToShortCar(car models.Car) models.ShortCar {
	return models.ShortCar{
		CarUID: car.CarUID,
		Brand: car.Brand,
		Model: car.Model,
		RegistrationNumber: car.RegistrationNumber,
		Availability: car.Availability,
	}
}