package service

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func TestTextGenerateToken(t *testing.T) {
	testCases := []struct {
		name   string
		userID int64
	}{
		{name: "positive user id", userID: 42},
		{name: "zero user id", userID: 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			authService := NewAuthService([]byte("secret"), time.Hour)

			token, err := authService.GenerateToken(tc.userID)
			if err != nil {
				t.Fatalf("GenerateToken returned error: %v", err)
			}
			if token == "" {
				t.Fatal("GenerateToken returned empty token")
			}

			parsedUserID, err := authService.ParseToken(token)
			if err != nil {
				t.Fatalf("ParseToken returned error: %v", err)
			}
			if parsedUserID != tc.userID {
				t.Fatalf("ParseToken userID = %d, want %d", parsedUserID, tc.userID)
			}
		})
	}
}

func TestTextParseToken(t *testing.T) {
	validService := NewAuthService([]byte("secret"), time.Hour)
	validToken, err := validService.GenerateToken(42)
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}

	wrongMethodToken, err := jwt.NewWithClaims(jwt.SigningMethodHS512, Claims{
		UserID: 42,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}).SignedString([]byte("secret"))
	if err != nil {
		t.Fatalf("SignedString returned error: %v", err)
	}

	testCases := []struct {
		name        string
		service     *AuthService
		token       string
		expectedID  int64
		expectedErr error
		errContains string
	}{
		{name: "valid token", service: validService, token: validToken, expectedID: 42},
		{name: "invalid token format", service: validService, token: "broken.token", errContains: "invalid number of segments"},
		{name: "wrong signing method", service: validService, token: wrongMethodToken, expectedErr: ErrInvalidToken},
		{name: "wrong secret", service: NewAuthService([]byte("other"), time.Hour), token: validToken, expectedErr: jwt.ErrSignatureInvalid},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			userID, err := tc.service.ParseToken(tc.token)
			if tc.errContains != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tc.errContains)
				}
				if !strings.Contains(err.Error(), tc.errContains) {
					t.Fatalf("unexpected error: got %v want substring %q", err, tc.errContains)
				}
				return
			}
			if tc.expectedErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tc.expectedErr)
				}
				if !errors.Is(err, tc.expectedErr) {
					t.Fatalf("unexpected error: got %v want %v", err, tc.expectedErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("ParseToken returned error: %v", err)
			}
			if userID != tc.expectedID {
				t.Fatalf("ParseToken userID = %d, want %d", userID, tc.expectedID)
			}
		})
	}
}
