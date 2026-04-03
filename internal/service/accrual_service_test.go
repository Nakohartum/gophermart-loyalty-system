package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTextNewAccrualService(t *testing.T) {
	testCases := []struct {
		name     string
		baseURL  string
		expected string
	}{
		{name: "without trailing slash", baseURL: "http://localhost:8080", expected: "http://localhost:8080"},
		{name: "with trailing slash", baseURL: "http://localhost:8080/", expected: "http://localhost:8080"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := NewAccrualService(tc.baseURL)
			if service.baseURL != tc.expected {
				t.Fatalf("baseURL = %q, want %q", service.baseURL, tc.expected)
			}
			if service.client == nil {
				t.Fatal("client was not initialized")
			}
		})
	}
}

func TestTextGetOrder(t *testing.T) {
	testCases := []struct {
		name           string
		statusCode     int
		responseBody   string
		expectedNil    bool
		expectedStatus string
		expectedErr    error
	}{
		{name: "ok", statusCode: http.StatusOK, responseBody: `{"order":"123","status":"PROCESSED","accrual":10.5}`, expectedStatus: "PROCESSED"},
		{name: "no content", statusCode: http.StatusNoContent, expectedNil: true},
		{name: "too many requests", statusCode: http.StatusTooManyRequests, expectedErr: ErrAccrualTryLater},
		{name: "unexpected status", statusCode: http.StatusInternalServerError, expectedErr: errUnexpectedAccrualStatus500()},
		{name: "invalid json", statusCode: http.StatusOK, responseBody: `{`, expectedErr: errUnexpectedJSON()},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/orders/123" {
					t.Fatalf("unexpected path: %s", r.URL.Path)
				}
				w.WriteHeader(tc.statusCode)
				_, _ = w.Write([]byte(tc.responseBody))
			}))
			defer server.Close()

			service := NewAccrualService(server.URL)
			order, err := service.GetOrder(context.Background(), "123")

			switch {
			case tc.expectedErr == ErrAccrualTryLater:
				if err != ErrAccrualTryLater {
					t.Fatalf("GetOrder error = %v, want %v", err, ErrAccrualTryLater)
				}
			case tc.name == "unexpected status":
				if err == nil || err.Error() != tc.expectedErr.Error() {
					t.Fatalf("GetOrder error = %v, want %v", err, tc.expectedErr)
				}
			case tc.name == "invalid json":
				if err == nil {
					t.Fatal("expected JSON decode error, got nil")
				}
			default:
				if err != nil {
					t.Fatalf("GetOrder returned error: %v", err)
				}
			}

			if tc.expectedNil {
				if order != nil {
					t.Fatalf("expected nil order, got %#v", order)
				}
				return
			}
			if tc.expectedErr != nil {
				if order != nil {
					t.Fatalf("expected nil order on error, got %#v", order)
				}
				return
			}

			if order == nil {
				t.Fatal("expected non-nil order")
			}
			if string(order.Status) != tc.expectedStatus {
				t.Fatalf("order status = %q, want %q", order.Status, tc.expectedStatus)
			}
		})
	}
}

func errUnexpectedAccrualStatus500() error {
	return &unexpectedAccrualStatusError{statusCode: http.StatusInternalServerError}
}

func errUnexpectedJSON() error {
	return &unexpectedJSONError{}
}

type unexpectedAccrualStatusError struct {
	statusCode int
}

func (e *unexpectedAccrualStatusError) Error() string {
	return "unexpected accrual status: 500"
}

type unexpectedJSONError struct{}

func (e *unexpectedJSONError) Error() string {
	return "unexpected EOF"
}
