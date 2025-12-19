package services

import (

)

type IGatewayService interface {

}

type Services struct {
	IGatewayService
}

func NewServices() *Services {
	return &Services{
		IGatewayService: NewGatewayService(),
	}
}