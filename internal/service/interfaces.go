package service

import (
	"context"

	"github.com/Nakohartum/gophermart-loyalty-system/internal/model"
)

type Database interface {
	RegisterUser(ctx context.Context, login, password string) (int64, error)
	AuthenticateUser(ctx context.Context, login, password string) (int64, error)
	CreateOrder(ctx context.Context, order model.Order, userID int64) (string, error)
	GetOrdersForProcessing(ctx context.Context) ([]model.Order, error)
	UpdateOrder(ctx context.Context, userID int64, number string, status model.Status, accrual *float64) error
	GetListOfUploadedOrders(ctx context.Context, userID int64) []model.OrderResponse
	GetCurrentUserBalance(ctx context.Context, userID int64) model.BalanceResponse
	WithdrawBalance(ctx context.Context, request model.WithdrawRequest) error
	GetWithdrawalsInfo(ctx context.Context, userID int64) []model.Withdrawal
}

type Auth interface {
	GenerateToken(userID int64) (string, error)
	ParseToken(tokenString string) (int64, error)
}

type Accrual interface {
	GetOrder(ctx context.Context, number string) (*model.AccrualOrder, error)
}
