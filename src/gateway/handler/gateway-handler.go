package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/SwanPoi/bmstu_rsoi_lab2/src/gateway/converters"
	"github.com/SwanPoi/bmstu_rsoi_lab2/src/gateway/models"
	cb "github.com/SwanPoi/bmstu_rsoi_lab2/src/gateway/circuitBreaker"
	queue "github.com/SwanPoi/bmstu_rsoi_lab2/src/gateway/queue"

	"github.com/gin-gonic/gin"
)

func forwardRequest(c *gin.Context, method, targetURL string, headers map[string]string, body []byte) (int, []byte, http.Header, error) {
	if len(c.Request.URL.RawQuery) > 0 {
		targetURL = fmt.Sprintf("%s?%s", targetURL, c.Request.URL.RawQuery)
	}

	req, err := http.NewRequest(method, targetURL, bytes.NewReader(body))
	if err != nil {
		return 0, nil, nil, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	if c.Request.Header.Get("Content-Type") != "" {
		req.Header.Set("Content-Type", c.Request.Header.Get("Content-Type"))
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, resp.Header, err
	}

	return resp.StatusCode, respBody, resp.Header, nil
}

// Функция forward с Circuit Breaker
func (h *GatewayHandler) forwardRequestWithCB(
	c *gin.Context,
	method, targetURL string,
	headers map[string]string,
	body []byte,
	cb *cb.CircuitBreaker,
	isCritical bool,
) (int, []byte, http.Header, error) {

	if !cb.AllowRequest() {
		if isCritical {
			log.Println(targetURL + " is unavailable (critical)")
			return 0, nil, nil, fmt.Errorf(targetURL + " is unavailable")
		}
		log.Println(targetURL + " is unavailable (not critical)")
		return http.StatusOK, []byte("{}"), nil, nil
	}

	if len(c.Request.URL.RawQuery) > 0 {
		targetURL = fmt.Sprintf("%s?%s", targetURL, c.Request.URL.RawQuery)
	}

	req, err := http.NewRequest(method, targetURL, bytes.NewReader(body))
	if err != nil {
		cb.RecordFailure()
		return 0, nil, nil, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	if c.Request.Header.Get("Content-Type") != "" {
		req.Header.Set("Content-Type", c.Request.Header.Get("Content-Type"))
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		cb.RecordFailure()
		return 0, nil, nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		cb.RecordFailure()
		return resp.StatusCode, nil, resp.Header, err
	}

	if resp.StatusCode >= 500 {
		cb.RecordFailure()
		if isCritical {
			return resp.StatusCode, respBody, resp.Header, fmt.Errorf("service error: %d", resp.StatusCode)
		}
		return http.StatusOK, []byte("{}"), nil, nil
	}

	cb.RecordSuccess()
	return resp.StatusCode, respBody, resp.Header, nil
}

// Rollbacks
func (h *GatewayHandler) rollbackCarBooking(ctx *gin.Context, carUID string) {
	carStatusUpsert := models.CarStatusUpsert{Availability: true}
	carStatusBytes, _ := json.Marshal(carStatusUpsert)
	forwardRequest(ctx, "PATCH", h.config.CarUrl + "/cars/" + carUID, nil, carStatusBytes)
}

func (h *GatewayHandler) rollbackPayment(ctx *gin.Context, paymentUID string) {
	paymentStatusUpsert := models.PaymentUpsert{Status: "CANCELED"}
	paymentStatusBytes, _ := json.Marshal(paymentStatusUpsert)
	forwardRequest(ctx, "PATCH", h.config.PaymentUrl + "/payment/" + paymentUID, nil, paymentStatusBytes)
}

func (h *GatewayHandler) rollbackRental(ctx *gin.Context, rentalUID string, headers map[string]string) {
	rentalStatusUpsert := models.RentalUpsert{Status: "CANCELED"}
	rentalStatusBytes, _ := json.Marshal(rentalStatusUpsert)

	rentalUrl := h.config.RentalUrl + "/rental/" + rentalUID
	forwardRequest(ctx, "PATCH", rentalUrl, headers, rentalStatusBytes)
}



// Main functions
func (h *GatewayHandler) GetCars(ctx *gin.Context) {
	status, body, headers, err := h.forwardRequestWithCB(ctx, "GET", h.config.CarUrl + "/cars", nil, nil, h.carCB, true)

	if err != nil {
		log.Println("GET /cars, ", err.Error())
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{ Message: "Car Service unavailable" })
		return
	}

	ctx.Data(status, headers.Get("Content-Type"), body)
}

func (h *GatewayHandler) GetUserRentals(ctx *gin.Context) {
	username := ctx.GetHeader("X-User-Name")
	if username == "" {
		log.Println("GET /rentals, Need X-User-Name for rentals")
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "X-User-Name header is required"})
		return
	}

	headers := map[string]string{"X-User-Name": username}

	// 1. Получить аренды
	status, body, _, err := h.forwardRequestWithCB(ctx, "GET", h.config.RentalUrl + "/rental", headers, nil, h.rentalCB, true)

	if err != nil {
		log.Println("GET /rentals, can't get rentals", err.Error())
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Rental Service unavailable"})
		return
	}

	if status != http.StatusOK {
		ctx.Data(status, "application/json", body)
		return
	}

	var rentals []models.RentalInfo
	if err := json.Unmarshal(body, &rentals); err != nil {
		log.Println("GET /rentals, rental parsing error, ", err.Error())
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Rental parsing error"})
		return
	}

	carUIDs := make([]string, len(rentals))
	paymentUIDs := make([]string, len(rentals))

	for i, rental := range rentals {
		carUIDs[i] = rental.CarUID
		paymentUIDs[i] = rental.PaymentUID
	}
	// 2. Получить автомобили
	carUrl := h.config.CarUrl + "/cars/query"
	carsRequest := models.CarsRequest{ UIDs: carUIDs }
	carReqBody, _ := json.Marshal(carsRequest)

	carStatus, carBody, _, err := h.forwardRequestWithCB(ctx, "POST", carUrl, nil, carReqBody, h.carCB, false)
	if err != nil {
		log.Println("GET /rentals, can't get cars, ", err.Error())
	}

	var cars []models.ShortCarResponse
	if err == nil && carStatus == http.StatusOK && len(carBody) > 2 && string(carBody) != "{}" {
		if err := json.Unmarshal(carBody, &cars); err != nil {
			log.Println("GET /rentals, car parsing error, ", err.Error())
			cars = nil // Будем использовать fallback
		}
	}
	
	// 3. Получить оплаты
	paymentUrl := h.config.PaymentUrl + "/payment/query"
	paymentsRequest := models.PaymentsRequest{ UIDs: paymentUIDs }
	paymentsReqBody, _ := json.Marshal(paymentsRequest)

	paymentStatus, paymentBody, _, err := h.forwardRequestWithCB(ctx, "POST", paymentUrl, nil, paymentsReqBody, h.paymentCB, false)
	if err != nil {
		log.Println("GET /rentals, can't get payments, ", err.Error())
	}

	var payments []models.PaymentInfo
	if err == nil && paymentStatus == http.StatusOK && len(paymentBody) > 2 && string(paymentBody) != "{}" {
		if err := json.Unmarshal(paymentBody, &payments); err != nil {
			log.Println("GET /rentals, payments parsing error, ", err.Error())
			payments = nil // Будем использовать fallback
		}
	}
	
	// 4. Смэтчить в массив RentalResponse
	carMap := make(map[string]models.CarInfo)
	if cars != nil {
		for _, car := range cars {
			carMap[car.CarUID] = models.CarInfo{
				CarUID:            car.CarUID,
				Brand:             car.Brand,
				Model:             car.Model,
				RegistrationNumber: car.RegistrationNumber,
			}
		}
	}

	for _, rental := range rentals {
		if _, exists := carMap[rental.CarUID]; !exists {
			carMap[rental.CarUID] = models.CarInfo{CarUID: rental.CarUID}
		}
	}

    paymentMap := make(map[string]models.PaymentInfo)
	if payments != nil {
		for _, payment := range payments {
			paymentMap[payment.PaymentUID] = payment
		}
	}

	for _, rental := range rentals {
		if _, exists := paymentMap[rental.PaymentUID]; !exists {
			paymentMap[rental.PaymentUID] = models.PaymentInfo{PaymentUID: rental.PaymentUID}
		}
	}

	rentalsResponse := make([]models.RentalResponse, len(rentals))

	for i, rental := range rentals {
		rentalsResponse[i] = converters.ConvertToRentalResponse(rental, carMap[rental.CarUID], paymentMap[rental.PaymentUID])
	}

	ctx.JSON(http.StatusOK, rentalsResponse)
}

func (h *GatewayHandler) GetRentalById(ctx *gin.Context) {
	username := ctx.GetHeader("X-User-Name")
	if username == "" {
		log.Println("GET /rental/:id, Need X-User-Name for rental")
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "X-User-Name header is required"})
		return
	}

	headers := map[string]string{"X-User-Name": username}

	rentalUid := ctx.Param("rentalUid")

	if rentalUid == "" {
		log.Println("GET /rental/:id, need valid uid")
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "RentalUid is required"})
	}

	rentalUrl := h.config.RentalUrl + "/rental/" + rentalUid
	
	status, body, _, err := h.forwardRequestWithCB(ctx, "GET", rentalUrl, headers, nil, h.rentalCB, true)

	if err != nil {
		log.Println("GET /rental/:id, can't get rental with id = " + rentalUid + ", ", err.Error())
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{ Message: "Rental Service unavailable" })
		return
	}

	if status != http.StatusOK {
		log.Println("GET /rental/:id, rental getting error with uid = " + rentalUid)
		ctx.Data(status, "application/json", body)
		return
	}

	var rental models.RentalInfo
	if err := json.Unmarshal(body, &rental); err != nil {
		log.Println("GET /rental/:id, rental parsing error")
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Rental parsing error"})
		return
	}

	// Получить авто
	carUrl := h.config.CarUrl + "/cars/" + rental.CarUID

	carStatus, carBody, _, err := h.forwardRequestWithCB(ctx, "GET", carUrl, nil, nil, h.carCB, false)
	// if err != nil {
	// 	log.Println("GET /rental/:id, can't get car with uid = " + rental.CarUID + " ", err.Error())
	// 	ctx.JSON(http.StatusBadGateway, models.ErrorResponse{Message: err.Error()})
	// 	return
	// }

	// TODO: уточнить
	if carStatus != http.StatusOK && err == nil {
		log.Println("GET /rental/:id, car getting error with uid = " + rental.CarUID + " ")
		ctx.Data(carStatus, "application/json", carBody)
		return
	}

	var car models.CarInfo
	if err == nil && carStatus == http.StatusOK && len(carBody) > 2 && string(carBody) != "{}" {
		var carResponse models.ShortCarResponse
		if err := json.Unmarshal(carBody, &carResponse); err == nil {
			car = models.CarInfo{
				CarUID:            carResponse.CarUID,
				Brand:             carResponse.Brand,
				Model:             carResponse.Model,
				RegistrationNumber: carResponse.RegistrationNumber,
			}
		} else {
			car = models.CarInfo{CarUID: rental.CarUID}
		}
	} else {
		car = models.CarInfo{CarUID: rental.CarUID}
	}
	
	// Получить оплату
	paymentUrl := h.config.PaymentUrl + "/payment/" + rental.PaymentUID

	paymentStatus, paymentBody, _, err := h.forwardRequestWithCB(ctx, "GET", paymentUrl, nil, nil, h.paymentCB, false)
	// if err != nil {
	// 	log.Println("GET /rental/:id, can't get payment with with uid = " + rental.PaymentUID + " ", err.Error())
	// 	ctx.JSON(http.StatusBadGateway, models.ErrorResponse{Message: err.Error()})
	// 	return
	// }

	if paymentStatus != http.StatusOK && err == nil {
		log.Println("GET /rental/:id, payment getting error with with uid = " + rental.CarUID + " ")
		ctx.Data(paymentStatus, "application/json", paymentBody)
		return
	}

	var payment models.PaymentInfo
	if err == nil && paymentStatus == http.StatusOK && len(paymentBody) > 2 && string(paymentBody) != "{}" {
		if err := json.Unmarshal(paymentBody, &payment); err != nil {
			payment = models.PaymentInfo{PaymentUID: rental.PaymentUID}
		}
	} else {
		payment = models.PaymentInfo{PaymentUID: rental.PaymentUID}
	}

	response := converters.ConvertToRentalResponse(rental, car, payment)

	ctx.JSON(http.StatusOK, response)

}

func (h *GatewayHandler) RentCar(ctx *gin.Context) {
	username := ctx.GetHeader("X-User-Name")
	if username == "" {
		log.Println("POST /rental, Need X-User-Name for car rent creation")
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "X-User-Name header is required"})
		return
	}

	bodyBytes, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		log.Println("POST /rental, invalid body")
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Fail during reading of request body for car rent"})
		return
	}

	ctx.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var rentReq models.RentCreationRequest
	if err := json.Unmarshal(bodyBytes, &rentReq); err != nil {
		log.Println("POST /rental, body parsing error, ", err.Error())
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Rent request parsing error"})
		return
	}

	checkCarUrl := h.config.CarUrl + "/cars/" + rentReq.CarUID

	carStatus, carBody, _, err := forwardRequest(ctx, "GET", checkCarUrl, nil, nil)
	if err != nil {
		log.Println("POST /rental, can't get car with uid = " + rentReq.CarUID + " ", err.Error())
		ctx.JSON(http.StatusServiceUnavailable, models.ErrorResponse{ Message: "Car Service unavailable" })
		return
	}

	if carStatus != http.StatusOK {
		log.Println("POST /rental, car getting error with uid = " + rentReq.CarUID + " ")
		ctx.Data(carStatus, "application/json", carBody)
		return
	}

	var carResponse models.ShortCarResponse
	if err := json.Unmarshal(carBody, &carResponse); err != nil {
		log.Println("POST /rental, car parsing error with uid = " + rentReq.CarUID)
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Car parsing error"})
		return
	}

	if !carResponse.Availability {
		log.Println("POST /rental, can't rent unavailable car with uid = " + rentReq.CarUID)
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "can't rent unavailable car with uid = " + rentReq.CarUID})
		return
	}

	carStatusUpsert := models.CarStatusUpsert{
		Availability: false,
	}

	carStatusBytes, err := json.Marshal(carStatusUpsert)

	if err != nil {
		log.Println("POST /rental, car request marshaling error, ", err.Error())
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Rent request parsing error"})
		return
	}

	carUrl := h.config.CarUrl + "/cars/" + rentReq.CarUID

	status, body, _, err := forwardRequest(ctx, "PATCH", carUrl, nil, carStatusBytes)

	if err != nil {
		log.Println("POST /rental, can't update car with uid = " + rentReq.CarUID + " ", err.Error())
		ctx.JSON(http.StatusServiceUnavailable, models.ErrorResponse{ Message: "Car Service unavailable" })
		return
	}

	if status != http.StatusOK {
		log.Println("POST /rental, updating error car with uid = " + rentReq.CarUID)
		ctx.Data(status, "application/json", body)
		return
	}

	payCreateReq := models.PaymentCreateRequest{
		DateFrom: rentReq.DateFrom,
		DateTo: rentReq.DateTo,
	}

	payCreateBytes, err := json.Marshal(payCreateReq)

	if err != nil {
		log.Println("POST /rental, payment request marshaling error, ", err.Error())
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Payment Creation request marshaling error"})
		return
	}

	payStatus, payBody, _, err := forwardRequest(ctx, "POST", h.config.PaymentUrl + "/payment", nil, payCreateBytes)

	if err != nil {
		log.Println("POST /rental, can't create payment, ", err.Error())
		h.rollbackCarBooking(ctx, rentReq.CarUID) // TODO: Queue
		ctx.JSON(http.StatusServiceUnavailable, models.ErrorResponse{Message: "Payment Service unavailable"})
		return
	}

	if payStatus != http.StatusOK {
		log.Println("POST /rental, payment creation error")
		h.rollbackCarBooking(ctx, rentReq.CarUID) // TODO: Queue
		ctx.Data(payStatus, "application/json", payBody)
		return
	}

	var paymentResponse models.PaymentCreationResponse

	if err := json.Unmarshal(payBody, &paymentResponse); err != nil {
		log.Println("POST /rental, payment parsing, ", err.Error())
		h.rollbackCarBooking(ctx, rentReq.CarUID) // TODO: Queue
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Payment response parsing error"})
		return
	}

	rentCreation := models.RentCreation{
		DateFrom: rentReq.DateFrom,
		DateTo: rentReq.DateTo,
		CarUID: rentReq.CarUID,
		PaymentUID: paymentResponse.PaymentUID,
		Username: username,
	}

	rentBytes, err := json.Marshal(rentCreation)

	if err != nil {
		log.Println("POST /rental, rental marshaling error, ", err.Error())
		h.rollbackCarBooking(ctx, rentReq.CarUID) // TODO: Queue
		h.rollbackPayment(ctx, paymentResponse.PaymentUID) // TODO: Queue
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Rental Creation request marshaling error"})
		return
	}

	rentStatus, rentBody, _, err := forwardRequest(ctx, "POST", h.config.RentalUrl + "/rental", nil, rentBytes)

	if err != nil {
		log.Println("POST /rental, can't create rental, ", err.Error())
		h.rollbackCarBooking(ctx, rentReq.CarUID) // TODO: Queue
		h.rollbackPayment(ctx, paymentResponse.PaymentUID) // TODO: Queue
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	if rentStatus != http.StatusOK {
		log.Println("POST /rental, rental creation error")
		h.rollbackCarBooking(ctx, rentReq.CarUID) // TODO: Queue
		h.rollbackPayment(ctx, paymentResponse.PaymentUID) // TODO: Queue
		ctx.Data(rentStatus, "application/json", rentBody)
		return
	}

	var rentalCreationResponse models.RentalInfo

	if err := json.Unmarshal(rentBody, &rentalCreationResponse); err != nil {
		log.Println("POST /rental, can't parse rental, ", err.Error())
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Rental Creation response parsing error"})
		return
	}

	rentResponse := converters.ConvertToCreateRentalResponse(rentalCreationResponse, paymentResponse)

	ctx.JSON(http.StatusOK, rentResponse)
}

func (h *GatewayHandler) FinishCarRent(ctx *gin.Context) {
	username := ctx.GetHeader("X-User-Name")
	if username == "" {
		log.Println("Need X-User-Name for car rent finishing")
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "X-User-Name header is required"})
		return
	}

	headers := map[string]string{"X-User-Name": username}

	rentalUid := ctx.Param("rentalUid")

	if rentalUid == "" {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "RentalUid is required"})
	}

	checkRentalUrl := h.config.RentalUrl + "/rental/" + rentalUid
	
	status, body, _, err := forwardRequest(ctx, "GET", checkRentalUrl, headers, nil)

	if err != nil {
		log.Println("POST /rental/:id/finish, can't get rental with id = " + rentalUid + ", ", err.Error())
		ctx.JSON(http.StatusServiceUnavailable, models.ErrorResponse{Message: "Rental Service unavailable"})
		return
	}

	if status != http.StatusOK {
		log.Println("POST /rental/:id/finish, rental getting error with uid = " + rentalUid)
		ctx.Data(status, "application/json", body)
		return
	}

	var rental models.RentalInfo
	if err := json.Unmarshal(body, &rental); err != nil {
		log.Println("POST /rental/:id/finish, rental parsing error")
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Rental parsing error"})
		return
	}

	if rental.Status != "IN_PROGRESS" {
		log.Println("POST /rental/:id/finish, rental with id = ", rental.RentalUID, " is not active")
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Rental with id = " + rental.RentalUID + " is not active"})
		return
	}

	carStatusUpsert := models.CarStatusUpsert{
		Availability: true,
	}

	carStatusBytes, err := json.Marshal(carStatusUpsert)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Car upsert marshalling error"})
		return
	}

	carUrl := h.config.CarUrl + "/cars/" + rental.CarUID

	carStatus, _, _, err := forwardRequest(ctx, "PATCH", carUrl, nil, carStatusBytes)
	if err != nil || carStatus != http.StatusOK {
		queue.EnqueueRetry(queue.RetryRequest{
			Method:  "PATCH",
			URL:    carUrl,
			Headers: nil,
			Body:    carStatusBytes,
		})
	}

	rentalReq := models.RentalUpsert{
		Status: "FINISHED",
	}

	rentalBytes, err := json.Marshal(rentalReq)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Payment Creation request marshaling error"})
		return
	}

	rentalUrl := h.config.RentalUrl + "/rental/" + rentalUid
	
	status, rentBody, _, err := forwardRequest(ctx, "PATCH", rentalUrl, headers, rentalBytes)

	if err != nil || status != http.StatusOK {
		queue.EnqueueRetry(queue.RetryRequest{
			Method:  "PATCH",
			URL:    rentalUrl,
			Headers: headers,
			Body:    rentalBytes,
		})
	}

	var rentalResponse models.RentalInfo

	if err := json.Unmarshal(rentBody, &rentalResponse); err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Rental Deletion response parsing error"})
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (h *GatewayHandler) RevokeRent(ctx *gin.Context) {
	username := ctx.GetHeader("X-User-Name")
	if username == "" {
		log.Println("Need X-User-Name for rental revoking")
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "X-User-Name header is required"})
		return
	}

	headers := map[string]string{"X-User-Name": username}

	rentalUid := ctx.Param("rentalUid")

	if rentalUid == "" {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "RentalUid is required"})
	}

	checkRentalUrl := h.config.RentalUrl + "/rental/" + rentalUid
	
	status, body, _, err := forwardRequest(ctx, "GET", checkRentalUrl, headers, nil)

	if err != nil {
		log.Println("DELETE /rental/:id, can't get rental with id = " + rentalUid + ", ", err.Error())
		ctx.JSON(http.StatusServiceUnavailable, models.ErrorResponse{Message: "Rental Service unavailable"})
		return
	}

	if status != http.StatusOK {
		log.Println("DELETE /rental/:id, rental getting error with uid = " + rentalUid)
		ctx.Data(status, "application/json", body)
		return
	}

	var rental models.RentalInfo
	if err := json.Unmarshal(body, &rental); err != nil {
		log.Println("DELETE /rental/:id, rental parsing error")
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Rental parsing error"})
		return
	}

	if rental.Status != "IN_PROGRESS" {
		log.Println("DELETE /rental/:id, rental with id = ", rental.RentalUID, " is not active")
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Rental with id = " + rental.RentalUID + " is not active"})
		return
	}

	carStatusUpsert := models.CarStatusUpsert{
		Availability: true,
	}

	carStatusBytes, err := json.Marshal(carStatusUpsert)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Car upsert marshalling error"})
		return
	}

	carUrl := h.config.CarUrl + "/cars/" + rental.CarUID

	carStatus, _, _, err := forwardRequest(ctx, "PATCH", carUrl, nil, carStatusBytes)

	if err != nil || carStatus != http.StatusOK {
		queue.EnqueueRetry(queue.RetryRequest{
			Method:  "PATCH",
			URL:     carUrl,
			Headers: nil,
			Body:    carStatusBytes,
		})
		log.Printf("Car upsert queued for retry: %s", rentalUid)
	}

	rentalReq := models.RentalUpsert{
		Status: "CANCELED",
	}

	rentalBytes, err := json.Marshal(rentalReq)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Rental request marshaling error"})
		return
	}

	rentalUrl := h.config.RentalUrl + "/rental/" + rentalUid
	
	status, rentBody, _, err := forwardRequest(ctx, "PATCH", rentalUrl, headers, rentalBytes)

	if err != nil || status != http.StatusOK  {
		queue.EnqueueRetry(queue.RetryRequest{
			Method:  "PATCH",
			URL:    rentalUrl,
			Headers: headers,
			Body:    rentalBytes,
		})
	}

	var rentalResponse models.RentalInfo

	if err := json.Unmarshal(rentBody, &rentalResponse); err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Rental Deletion response parsing error"})
		return
	}

	paymentStatusUpsert := models.PaymentUpsert{
		Status: "CANCELED",
	}

	paymentStatusBytes, err := json.Marshal(paymentStatusUpsert)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Payment upsert marshalling error"})
		return
	}

	paymentUrl := h.config.PaymentUrl + "/payment/" + rentalResponse.PaymentUID

	paymentStatus, _, _, err := forwardRequest(ctx, "PATCH", paymentUrl, nil, paymentStatusBytes)
	if err != nil || paymentStatus != http.StatusOK {
		queue.EnqueueRetry(queue.RetryRequest{
			Method:  	"PATCH",
			URL:    	paymentUrl,
			Headers: 	nil,
			Body:    	paymentStatusBytes,
		})
	}

	ctx.Status(http.StatusNoContent)
}