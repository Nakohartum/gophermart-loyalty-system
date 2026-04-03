package handler

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Nakohartum/gophermart-loyalty-system/internal/apperrors"
	"github.com/Nakohartum/gophermart-loyalty-system/internal/mocks"
	"github.com/Nakohartum/gophermart-loyalty-system/internal/model"
	"github.com/Nakohartum/gophermart-loyalty-system/internal/service"
	"go.uber.org/mock/gomock"
)

func TestTextCreateOrder(t *testing.T) {
	testCases := []struct {
		name               string
		method             string
		body               string
		withUser           bool
		setup              func(repo *mocks.MockRepository)
		expectedStatusCode int
	}{
		{name: "wrong method", method: http.MethodGet, expectedStatusCode: http.StatusBadRequest},
		{name: "missing user", method: http.MethodPost, body: "79927398713", expectedStatusCode: http.StatusUnauthorized},
		{name: "empty body", method: http.MethodPost, body: " ", withUser: true, expectedStatusCode: http.StatusBadRequest},
		{name: "invalid luhn", method: http.MethodPost, body: "123", withUser: true, expectedStatusCode: http.StatusUnprocessableEntity},
		{
			name:     "already uploaded by user",
			method:   http.MethodPost,
			body:     "79927398713",
			withUser: true,
			setup: func(repo *mocks.MockRepository) {
				repo.EXPECT().CreateOrder(gomock.Any(), gomock.Any(), int64(42)).Return("", apperrors.ErrOrderAlreadyUploadedByUser)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:     "already uploaded by another user",
			method:   http.MethodPost,
			body:     "79927398713",
			withUser: true,
			setup: func(repo *mocks.MockRepository) {
				repo.EXPECT().CreateOrder(gomock.Any(), gomock.Any(), int64(42)).Return("", apperrors.ErrOrderAlreadyUploadedByAnotherUser)
			},
			expectedStatusCode: http.StatusConflict,
		},
		{
			name:     "database error",
			method:   http.MethodPost,
			body:     "79927398713",
			withUser: true,
			setup: func(repo *mocks.MockRepository) {
				repo.EXPECT().CreateOrder(gomock.Any(), gomock.Any(), int64(42)).Return("", errors.New("boom"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			name:     "success",
			method:   http.MethodPost,
			body:     "79927398713",
			withUser: true,
			setup: func(repo *mocks.MockRepository) {
				repo.EXPECT().CreateOrder(gomock.Any(), gomock.AssignableToTypeOf(model.Order{}), int64(42)).DoAndReturn(
					func(ctx context.Context, order model.Order, userID int64) (string, error) {
						if order.Number != "79927398713" || order.Status != model.NEW || userID != 42 {
							t.Fatalf("unexpected order payload: %#v userID=%d", order, userID)
						}
						if time.Since(order.UploadedAt) > time.Second {
							t.Fatalf("UploadedAt is too old: %v", order.UploadedAt)
						}
						return order.Number, nil
					},
				)
			},
			expectedStatusCode: http.StatusAccepted,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := mocks.NewMockRepository(ctrl)
			auth := mocks.NewMockAuth(ctrl)
			if tc.setup != nil {
				tc.setup(repo)
			}

			appService := service.NewAppService(repo, auth)
			handler := NewOrderHandler(appService)
			req := httptest.NewRequest(tc.method, "/api/user/orders", bytes.NewBufferString(tc.body))
			if tc.withUser {
				req = req.WithContext(context.WithValue(req.Context(), userIDContextKey, int64(42)))
			}
			recorder := httptest.NewRecorder()

			handler.CreateOrder().ServeHTTP(recorder, req)

			if recorder.Code != tc.expectedStatusCode {
				t.Fatalf("status code = %d, want %d", recorder.Code, tc.expectedStatusCode)
			}
		})
	}
}

func TestTextGetListOfUploadedOrders(t *testing.T) {
	testCases := []struct {
		name               string
		method             string
		withUser           bool
		setup              func(repo *mocks.MockRepository)
		expectedStatusCode int
		expectedBody       string
	}{
		{name: "wrong method", method: http.MethodPost, expectedStatusCode: http.StatusMethodNotAllowed},
		{name: "missing user", method: http.MethodGet, expectedStatusCode: http.StatusUnauthorized},
		{
			name:     "no orders",
			method:   http.MethodGet,
			withUser: true,
			setup: func(repo *mocks.MockRepository) {
				repo.EXPECT().GetListOfUploadedOrders(gomock.Any(), int64(42)).Return(nil)
			},
			expectedStatusCode: http.StatusNoContent,
		},
		{
			name:     "success",
			method:   http.MethodGet,
			withUser: true,
			setup: func(repo *mocks.MockRepository) {
				value := 10.5
				repo.EXPECT().GetListOfUploadedOrders(gomock.Any(), int64(42)).Return([]model.OrderResponse{
					{Number: "79927398713", Status: model.PROCESSED, Accrual: &value},
				})
			},
			expectedStatusCode: http.StatusOK,
			expectedBody:       "[{\"number\":\"79927398713\",\"status\":\"PROCESSED\",\"accrual\":10.5,\"uploaded_at\":\"0001-01-01T00:00:00Z\"}]\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := mocks.NewMockRepository(ctrl)
			auth := mocks.NewMockAuth(ctrl)
			if tc.setup != nil {
				tc.setup(repo)
			}

			appService := service.NewAppService(repo, auth)
			handler := NewOrderHandler(appService)
			req := httptest.NewRequest(tc.method, "/api/user/orders", nil)
			if tc.withUser {
				req = req.WithContext(context.WithValue(req.Context(), userIDContextKey, int64(42)))
			}
			recorder := httptest.NewRecorder()

			handler.GetListOfUploadedOrders().ServeHTTP(recorder, req)

			if recorder.Code != tc.expectedStatusCode {
				t.Fatalf("status code = %d, want %d", recorder.Code, tc.expectedStatusCode)
			}
			if tc.expectedBody != "" && recorder.Body.String() != tc.expectedBody {
				t.Fatalf("body = %q, want %q", recorder.Body.String(), tc.expectedBody)
			}
		})
	}
}
