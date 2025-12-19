package converters

import "github.com/SwanPoi/bmstu_rsoi_lab2/src/car/models"

func CarResponseFromCar(car models.Car) models.CarResponse {
    return models.CarResponse{
        CarUID:           car.CarUID,
        Brand:            car.Brand,
        Model:            car.Model,
        RegistrationNumber: car.RegistrationNumber,
        Power:            car.Power,
        Type:             car.Type,
        Price:            car.Price,
        Available:        car.Availability,
    }
}

func CarResponsesFromCars(cars []models.Car) []models.CarResponse {
    responses := make([]models.CarResponse, len(cars))
    for i, car := range cars {
        responses[i] = CarResponseFromCar(car)
    }
    return responses
}