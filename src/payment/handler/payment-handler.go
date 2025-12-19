package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/SwanPoi/bmstu_rsoi_lab2/src/payment/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

/**
* Информация об оплате по идентификатору
 */
func (h *PaymentHandler) GetPaymentByUid(ctx *gin.Context) {
	paymentUid := ctx.Param("uid")

	if _, err := uuid.Parse(paymentUid); err != nil {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "PaymentUid must be valid"})
		return
	}

	payment, err := h.services.GetPaymentByUid(paymentUid)

	if err != nil {
		if errors.Is(err, models.ErrorNotFound) {
			message := "Payment with payment_uid = " + paymentUid + " is not found"
			ctx.JSON(http.StatusNotFound, models.ErrorResponse{Message: message})
		} else {
			ctx.JSON(http.StatusInternalServerError, err)
		}
		return
	}

	ctx.JSON(http.StatusOK, payment)
}

/**
* Получение оплат по идентификаторам
 */
func (h *PaymentHandler) GetPaymensBatch(ctx *gin.Context) {
	var req models.PaymentsRequest

	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Bad payments request model"})
		return
	}

	payments, err := h.services.GetPaymentsByUids(req.UIDs)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, payments)
}

func (h *PaymentHandler) UpdatePayment(ctx *gin.Context) {
	paymentUid := ctx.Param("uid")

	if _, err := uuid.Parse(paymentUid); err != nil {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Payment Uid must be valid"})
		return
	}

	var req models.PaymentUpsert

	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Bad Payment Upsert body"})
		return
	}

	payment, err := h.services.UpdatePayment(req, paymentUid)

	if err != nil {
		if err == models.ErrorNotFound {
			message := "Payment with uid = " + paymentUid + " is not found"
			ctx.JSON(http.StatusNotFound, models.ErrorResponse{Message: message})
		} else {
			ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		}

		return
	}

	ctx.JSON(http.StatusOK, payment)
}

func (h *PaymentHandler) CreatePayment(ctx *gin.Context) {
	var req models.PaymentCreateRequest

	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Bad Payment Creation body"})
		return
	}

	dateFrom, err := time.Parse("2006-01-02", req.DateFrom)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Error with parsing time from date-from"})
		return
	}

	dateTo, err := time.Parse("2006-01-02", req.DateTo)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Error with parsing time from date-from"})
		return
	}

	payment, err := h.services.CreatePayment(models.PaymentCreate{
		DateFrom: dateFrom,
		DateTo:   dateTo,
	})

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, payment)
}
