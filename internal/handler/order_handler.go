package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Nakohartum/gophermart-loyalty-system/internal/apperrors"
	"github.com/Nakohartum/gophermart-loyalty-system/internal/service"
)

type OrderHandler struct {
	orderService service.Orders
}

func NewOrderHandler(orderService service.Orders) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

func (oh *OrderHandler) CreateOrder() http.HandlerFunc {
	fun := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "not correct method", http.StatusBadRequest)
			return
		}
		userID, ok := UserIDFromContext(r.Context())

		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		body, err := io.ReadAll(r.Body)

		if err != nil {
			http.Error(w, "failed to read body", http.StatusBadRequest)
			return
		}

		if strings.TrimSpace(string(body)) == "" {
			http.Error(w, "empty order number", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		err = oh.orderService.CreateOrder(ctx, userID, string(body))

		if err != nil {
			if errors.Is(err, apperrors.ErrInvalidOrderNumber) {
				http.Error(w, "not correct number", http.StatusUnprocessableEntity)
				return
			}
			if errors.Is(err, apperrors.ErrOrderAlreadyUploadedByUser) {
				w.WriteHeader(http.StatusOK)
				return
			}
			if errors.Is(err, apperrors.ErrOrderAlreadyUploadedByAnotherUser) {
				http.Error(w, "uploaded by another user", http.StatusConflict)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusAccepted)
	}

	return fun
}

func (oh *OrderHandler) GetListOfUploadedOrders() http.HandlerFunc {
	fun := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "not correct method", http.StatusMethodNotAllowed)
			return
		}
		userID, ok := UserIDFromContext(r.Context())

		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		orders := oh.orderService.GetListOfUploadedOrders(ctx, userID)

		if len(orders) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(orders); err != nil {
			http.Error(w, "failed to encode data", http.StatusInternalServerError)
			return
		}
	}
	return fun
}
