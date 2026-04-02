package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Nakohartum/gophermart-loyalty-system/internal/config/db"
	"github.com/Nakohartum/gophermart-loyalty-system/internal/model"
	"github.com/Nakohartum/gophermart-loyalty-system/internal/service"
)

type OrderHandler struct {
	databaseService service.Database
	authService     service.Auth
}

func NewOrderHandler(dbService service.Database, authService service.Auth) *OrderHandler {
	return &OrderHandler{
		databaseService: dbService,
		authService:     authService,
	}
}

func (oh *OrderHandler) Orders() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			oh.CreateOrder()(w, r)
		case http.MethodGet:
			oh.GetListOfUploadedOrders()(w, r)
		default:
			http.Error(w, "not correct method", http.StatusMethodNotAllowed)
		}
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

		orderNumber := strings.TrimSpace(string(body))

		if orderNumber == "" {
			http.Error(w, "empty order number", http.StatusBadRequest)
			return
		}

		if !service.IsValidLuhn(orderNumber) {
			http.Error(w, "not correct number", http.StatusUnprocessableEntity)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)

		defer cancel()

		order := model.Order{
			Number:     orderNumber,
			Status:     model.NEW,
			UploadedAt: time.Now(),
		}

		_, err = oh.databaseService.CreateOrder(ctx, order, userID)

		if err != nil {
			if errors.Is(err, db.ErrOrderAlreadyUploadedByUser) {
				w.WriteHeader(http.StatusOK)
				return
			}
			if errors.Is(err, db.ErrOrderAlreadyUploadedByAnotherUser) {
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

		orders := oh.databaseService.GetListOfUploadedOrders(ctx, userID)

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
