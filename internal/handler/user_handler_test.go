package handler

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Nakohartum/gophermart-loyalty-system/internal/config/db"
	"github.com/Nakohartum/gophermart-loyalty-system/internal/mocks"
	"github.com/Nakohartum/gophermart-loyalty-system/internal/model"
	"go.uber.org/mock/gomock"
)

func TestTextRegisterUser(t *testing.T) {
	testCases := []struct {
		name               string
		method             string
		body               string
		setup              func(database *mocks.MockDatabase, auth *mocks.MockAuth)
		expectedStatusCode int
		expectedAuthHeader string
	}{
		{name: "wrong method", method: http.MethodGet, expectedStatusCode: http.StatusBadRequest},
		{name: "invalid json", method: http.MethodPost, body: "{", expectedStatusCode: http.StatusBadRequest},
		{name: "missing credentials", method: http.MethodPost, body: `{"login":"","password":""}`, expectedStatusCode: http.StatusBadRequest},
		{
			name:   "user already exists",
			method: http.MethodPost,
			body:   `{"login":"user","password":"pass"}`,
			setup: func(database *mocks.MockDatabase, auth *mocks.MockAuth) {
				database.EXPECT().RegisterUser(gomock.Any(), "user", "pass").Return(int64(0), db.ErrUserAlreadyExists)
			},
			expectedStatusCode: http.StatusConflict,
		},
		{
			name:   "database error",
			method: http.MethodPost,
			body:   `{"login":"user","password":"pass"}`,
			setup: func(database *mocks.MockDatabase, auth *mocks.MockAuth) {
				database.EXPECT().RegisterUser(gomock.Any(), "user", "pass").Return(int64(0), errors.New("boom"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			name:   "token error",
			method: http.MethodPost,
			body:   `{"login":"user","password":"pass"}`,
			setup: func(database *mocks.MockDatabase, auth *mocks.MockAuth) {
				database.EXPECT().RegisterUser(gomock.Any(), "user", "pass").Return(int64(42), nil)
				auth.EXPECT().GenerateToken(int64(42)).Return("", errors.New("token failed"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			name:   "success",
			method: http.MethodPost,
			body:   `{"login":"user","password":"pass"}`,
			setup: func(database *mocks.MockDatabase, auth *mocks.MockAuth) {
				database.EXPECT().RegisterUser(gomock.Any(), "user", "pass").Return(int64(42), nil)
				auth.EXPECT().GenerateToken(int64(42)).Return("token", nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedAuthHeader: "Bearer token",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			database := mocks.NewMockDatabase(ctrl)
			auth := mocks.NewMockAuth(ctrl)
			if tc.setup != nil {
				tc.setup(database, auth)
			}

			handler := NewUserHandler(database, auth)
			req := httptest.NewRequest(tc.method, "/api/user/register", bytes.NewBufferString(tc.body))
			recorder := httptest.NewRecorder()

			handler.RegisterUser().ServeHTTP(recorder, req)

			if recorder.Code != tc.expectedStatusCode {
				t.Fatalf("status code = %d, want %d", recorder.Code, tc.expectedStatusCode)
			}
			if tc.expectedAuthHeader != "" && recorder.Header().Get("Authorization") != tc.expectedAuthHeader {
				t.Fatalf("Authorization = %q, want %q", recorder.Header().Get("Authorization"), tc.expectedAuthHeader)
			}
		})
	}
}

func TestTextAuthenticateUser(t *testing.T) {
	testCases := []struct {
		name               string
		method             string
		body               string
		setup              func(database *mocks.MockDatabase, auth *mocks.MockAuth)
		expectedStatusCode int
		expectedAuthHeader string
	}{
		{name: "wrong method", method: http.MethodGet, expectedStatusCode: http.StatusBadRequest},
		{name: "invalid json", method: http.MethodPost, body: "{", expectedStatusCode: http.StatusBadRequest},
		{name: "missing credentials", method: http.MethodPost, body: `{"login":"","password":""}`, expectedStatusCode: http.StatusBadRequest},
		{
			name:   "password mismatch",
			method: http.MethodPost,
			body:   `{"login":"user","password":"pass"}`,
			setup: func(database *mocks.MockDatabase, auth *mocks.MockAuth) {
				database.EXPECT().AuthenticateUser(gomock.Any(), "user", "pass").Return(int64(0), db.ErrPasswordNotMatch)
			},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:   "database error",
			method: http.MethodPost,
			body:   `{"login":"user","password":"pass"}`,
			setup: func(database *mocks.MockDatabase, auth *mocks.MockAuth) {
				database.EXPECT().AuthenticateUser(gomock.Any(), "user", "pass").Return(int64(0), errors.New("boom"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			name:   "token error",
			method: http.MethodPost,
			body:   `{"login":"user","password":"pass"}`,
			setup: func(database *mocks.MockDatabase, auth *mocks.MockAuth) {
				database.EXPECT().AuthenticateUser(gomock.Any(), "user", "pass").Return(int64(42), nil)
				auth.EXPECT().GenerateToken(int64(42)).Return("", errors.New("token failed"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			name:   "success",
			method: http.MethodPost,
			body:   `{"login":"user","password":"pass"}`,
			setup: func(database *mocks.MockDatabase, auth *mocks.MockAuth) {
				database.EXPECT().AuthenticateUser(gomock.Any(), "user", "pass").Return(int64(42), nil)
				auth.EXPECT().GenerateToken(int64(42)).Return("token", nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedAuthHeader: "Bearer token",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			database := mocks.NewMockDatabase(ctrl)
			auth := mocks.NewMockAuth(ctrl)
			if tc.setup != nil {
				tc.setup(database, auth)
			}

			handler := NewUserHandler(database, auth)
			req := httptest.NewRequest(tc.method, "/api/user/login", bytes.NewBufferString(tc.body))
			recorder := httptest.NewRecorder()

			handler.AuthenticateUser().ServeHTTP(recorder, req)

			if recorder.Code != tc.expectedStatusCode {
				t.Fatalf("status code = %d, want %d", recorder.Code, tc.expectedStatusCode)
			}
			if tc.expectedAuthHeader != "" && recorder.Header().Get("Authorization") != tc.expectedAuthHeader {
				t.Fatalf("Authorization = %q, want %q", recorder.Header().Get("Authorization"), tc.expectedAuthHeader)
			}
		})
	}
}

func TestTextGetCurrentUserBalance(t *testing.T) {
	testCases := []struct {
		name               string
		method             string
		withUser           bool
		setup              func(database *mocks.MockDatabase)
		expectedStatusCode int
		expectedBody       string
	}{
		{name: "wrong method", method: http.MethodPost, expectedStatusCode: http.StatusBadRequest},
		{name: "missing user", method: http.MethodGet, expectedStatusCode: http.StatusUnauthorized},
		{
			name:     "success",
			method:   http.MethodGet,
			withUser: true,
			setup: func(database *mocks.MockDatabase) {
				database.EXPECT().GetCurrentUserBalance(gomock.Any(), int64(42)).Return(model.BalanceResponse{Current: 100.5, Withdrawn: 10.5})
			},
			expectedStatusCode: http.StatusOK,
			expectedBody:       "{\"current\":100.5,\"withdrawn\":10.5}\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			database := mocks.NewMockDatabase(ctrl)
			auth := mocks.NewMockAuth(ctrl)
			if tc.setup != nil {
				tc.setup(database)
			}

			handler := NewUserHandler(database, auth)
			req := httptest.NewRequest(tc.method, "/api/user/balance", nil)
			if tc.withUser {
				req = req.WithContext(context.WithValue(req.Context(), userIDContextKey, int64(42)))
			}
			recorder := httptest.NewRecorder()

			handler.GetCurrentUserBalance().ServeHTTP(recorder, req)

			if recorder.Code != tc.expectedStatusCode {
				t.Fatalf("status code = %d, want %d", recorder.Code, tc.expectedStatusCode)
			}
			if tc.expectedBody != "" && recorder.Body.String() != tc.expectedBody {
				t.Fatalf("body = %q, want %q", recorder.Body.String(), tc.expectedBody)
			}
		})
	}
}

func TestTextWithdrawBalance(t *testing.T) {
	testCases := []struct {
		name               string
		method             string
		body               string
		withUser           bool
		setup              func(database *mocks.MockDatabase)
		expectedStatusCode int
	}{
		{name: "wrong method", method: http.MethodGet, expectedStatusCode: http.StatusBadRequest},
		{name: "missing user", method: http.MethodPost, body: `{"order":"79927398713","sum":10}`, expectedStatusCode: http.StatusUnauthorized},
		{name: "invalid json", method: http.MethodPost, body: "{", withUser: true, expectedStatusCode: http.StatusBadRequest},
		{name: "invalid payload", method: http.MethodPost, body: `{"order":"","sum":0}`, withUser: true, expectedStatusCode: http.StatusBadRequest},
		{
			name:     "insufficient funds",
			method:   http.MethodPost,
			body:     `{"order":"79927398713","sum":10}`,
			withUser: true,
			setup: func(database *mocks.MockDatabase) {
				database.EXPECT().WithdrawBalance(gomock.Any(), model.WithdrawRequest{UserID: 42, Order: "79927398713", Sum: 10}).Return(db.ErrInsufficientFunds)
			},
			expectedStatusCode: http.StatusPaymentRequired,
		},
		{
			name:     "database error",
			method:   http.MethodPost,
			body:     `{"order":"79927398713","sum":10}`,
			withUser: true,
			setup: func(database *mocks.MockDatabase) {
				database.EXPECT().WithdrawBalance(gomock.Any(), model.WithdrawRequest{UserID: 42, Order: "79927398713", Sum: 10}).Return(errors.New("boom"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			name:     "success",
			method:   http.MethodPost,
			body:     `{"order":"79927398713","sum":10}`,
			withUser: true,
			setup: func(database *mocks.MockDatabase) {
				database.EXPECT().WithdrawBalance(gomock.Any(), model.WithdrawRequest{UserID: 42, Order: "79927398713", Sum: 10}).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			database := mocks.NewMockDatabase(ctrl)
			auth := mocks.NewMockAuth(ctrl)
			if tc.setup != nil {
				tc.setup(database)
			}

			handler := NewUserHandler(database, auth)
			req := httptest.NewRequest(tc.method, "/api/user/balance/withdraw", bytes.NewBufferString(tc.body))
			if tc.withUser {
				req = req.WithContext(context.WithValue(req.Context(), userIDContextKey, int64(42)))
			}
			recorder := httptest.NewRecorder()

			handler.WithdrawBalance().ServeHTTP(recorder, req)

			if recorder.Code != tc.expectedStatusCode {
				t.Fatalf("status code = %d, want %d", recorder.Code, tc.expectedStatusCode)
			}
		})
	}
}

func TestTextGetWithdrawalsInfo(t *testing.T) {
	testCases := []struct {
		name               string
		method             string
		withUser           bool
		setup              func(database *mocks.MockDatabase)
		expectedStatusCode int
		expectedBody       string
	}{
		{name: "wrong method", method: http.MethodPost, expectedStatusCode: http.StatusBadRequest},
		{name: "missing user", method: http.MethodGet, expectedStatusCode: http.StatusUnauthorized},
		{
			name:     "success",
			method:   http.MethodGet,
			withUser: true,
			setup: func(database *mocks.MockDatabase) {
				database.EXPECT().GetWithdrawalsInfo(gomock.Any(), int64(42)).Return([]model.Withdrawal{
					{Order: "79927398713", Sum: 10, ProcessedAt: time.Unix(0, 0).UTC()},
				})
			},
			expectedStatusCode: http.StatusOK,
			expectedBody:       "[{\"order\":\"79927398713\",\"sum\":10,\"processed_at\":\"1970-01-01T00:00:00Z\"}]\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			database := mocks.NewMockDatabase(ctrl)
			auth := mocks.NewMockAuth(ctrl)
			if tc.setup != nil {
				tc.setup(database)
			}

			handler := NewUserHandler(database, auth)
			req := httptest.NewRequest(tc.method, "/api/user/withdrawals", nil)
			if tc.withUser {
				req = req.WithContext(context.WithValue(req.Context(), userIDContextKey, int64(42)))
			}
			recorder := httptest.NewRecorder()

			handler.GetWithdrawalsInfo().ServeHTTP(recorder, req)

			if recorder.Code != tc.expectedStatusCode {
				t.Fatalf("status code = %d, want %d", recorder.Code, tc.expectedStatusCode)
			}
			if tc.expectedBody != "" && recorder.Body.String() != tc.expectedBody {
				t.Fatalf("body = %q, want %q", recorder.Body.String(), tc.expectedBody)
			}
		})
	}
}
