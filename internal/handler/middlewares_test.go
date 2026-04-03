package handler

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Nakohartum/gophermart-loyalty-system/internal/mocks"
	"go.uber.org/mock/gomock"
)

func TestTextGzipMiddleware(t *testing.T) {
	testCases := []struct {
		name               string
		requestBody        []byte
		contentEncoding    string
		acceptEncoding     string
		expectedStatusCode int
		expectedBody       string
		expectedCompressed bool
	}{
		{
			name:               "compresses response and decompresses request",
			requestBody:        gzipBytes(t, []byte("hello")),
			contentEncoding:    "gzip",
			acceptEncoding:     "gzip",
			expectedStatusCode: http.StatusOK,
			expectedBody:       "hello",
			expectedCompressed: true,
		},
		{
			name:               "passes plain request without compression",
			requestBody:        []byte("plain"),
			expectedStatusCode: http.StatusOK,
			expectedBody:       "plain",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatalf("ReadAll returned error: %v", err)
				}
				_, _ = w.Write(body)
			})

			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(tc.requestBody))
			if tc.contentEncoding != "" {
				req.Header.Set("Content-Encoding", tc.contentEncoding)
			}
			if tc.acceptEncoding != "" {
				req.Header.Set("Accept-Encoding", tc.acceptEncoding)
			}
			recorder := httptest.NewRecorder()

			GzipMiddleware(next).ServeHTTP(recorder, req)

			if recorder.Code != tc.expectedStatusCode {
				t.Fatalf("status code = %d, want %d", recorder.Code, tc.expectedStatusCode)
			}

			responseBody := recorder.Body.Bytes()
			if tc.expectedCompressed {
				if recorder.Header().Get("Content-Encoding") != "gzip" {
					t.Fatalf("Content-Encoding = %q, want gzip", recorder.Header().Get("Content-Encoding"))
				}
				responseBody = ungzipBytes(t, responseBody)
			}

			if string(responseBody) != tc.expectedBody {
				t.Fatalf("response body = %q, want %q", string(responseBody), tc.expectedBody)
			}
		})
	}
}

func TestTextAuthMiddleware(t *testing.T) {
	testCases := []struct {
		name               string
		authHeader         string
		setup              func(auth *mocks.MockAuth)
		expectedStatusCode int
		expectedUserID     int64
	}{
		{
			name:               "missing authorization header",
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "invalid bearer prefix",
			authHeader:         "Token abc",
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:       "parse token error",
			authHeader: "Bearer abc",
			setup: func(auth *mocks.MockAuth) {
				auth.EXPECT().ParseToken("abc").Return(int64(0), errors.New("bad token"))
			},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:       "success",
			authHeader: "Bearer abc",
			setup: func(auth *mocks.MockAuth) {
				auth.EXPECT().ParseToken("abc").Return(int64(42), nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedUserID:     42,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			auth := mocks.NewMockAuth(ctrl)
			if tc.setup != nil {
				tc.setup(auth)
			}

			var gotUserID int64
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotUserID, _ = UserIDFromContext(r.Context())
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}
			recorder := httptest.NewRecorder()

			AuthMiddleware(auth)(next).ServeHTTP(recorder, req)

			if recorder.Code != tc.expectedStatusCode {
				t.Fatalf("status code = %d, want %d", recorder.Code, tc.expectedStatusCode)
			}
			if tc.expectedStatusCode == http.StatusOK && gotUserID != tc.expectedUserID {
				t.Fatalf("userID = %d, want %d", gotUserID, tc.expectedUserID)
			}
		})
	}
}

func TestTextUserIDFromContext(t *testing.T) {
	testCases := []struct {
		name       string
		ctx        context.Context
		expectedID int64
		expectedOK bool
	}{
		{name: "missing value", ctx: context.Background(), expectedOK: false},
		{name: "present value", ctx: context.WithValue(context.Background(), userIDContextKey, int64(77)), expectedID: 77, expectedOK: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			userID, ok := UserIDFromContext(tc.ctx)
			if ok != tc.expectedOK || userID != tc.expectedID {
				t.Fatalf("UserIDFromContext() = (%d, %v), want (%d, %v)", userID, ok, tc.expectedID, tc.expectedOK)
			}
		})
	}
}

func gzipBytes(t *testing.T, body []byte) []byte {
	t.Helper()

	var buffer bytes.Buffer
	writer := gzip.NewWriter(&buffer)
	if _, err := writer.Write(body); err != nil {
		t.Fatalf("Write returned error: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}
	return buffer.Bytes()
}

func ungzipBytes(t *testing.T, body []byte) []byte {
	t.Helper()

	reader, err := gzip.NewReader(bytes.NewReader(body))
	if err != nil {
		t.Fatalf("NewReader returned error: %v", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("ReadAll returned error: %v", err)
	}
	return data
}
