package service

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type AuthService struct {
	secretKey []byte
	tokenTTL  time.Duration
}

func NewAuthService(secretKey []byte, tokenTTL time.Duration) *AuthService {
	return &AuthService{
		secretKey: secretKey,
		tokenTTL:  tokenTTL,
	}
}

type Claims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

func (authService *AuthService) GenerateToken(userID int64) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(authService.tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(authService.secretKey)

	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (authService *AuthService) ParseToken(tokenString string) (int64, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidToken
		}
		return authService.secretKey, nil
	})

	if err != nil {
		return 0, err
	}

	claims, ok := token.Claims.(*Claims)

	if !ok || !token.Valid {
		return 0, ErrInvalidToken
	}
	return claims.UserID, nil
}
