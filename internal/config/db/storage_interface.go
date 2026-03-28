package db

import (
	"context"
	"github.com/Nakohartum/gophermart-loyalty-system/internal/model"
)

type Storage interface {
	OpenConnection(ctx context.Context, databaseUri string) error
	CloseConnection(ctx context.Context) error
	CheckConnection(ctx context.Context) error
	RegisterUser(ctx context.Context, login, password string)(int64, error)
	AuthenticateUser(ctx context.Context, login, password string) (int64, error)
	CreateOrder(ctx context.Context, order model.Order, userId int64) (string, error)
	GetOrdersForProcessing(ctx context.Context) ([]model.Order, error)
	UpdateOrder(ctx context.Context, userId int64, number string, status model.Status, accrual *float64) error
	GetListOfUploadedOrders(ctx context.Context) []model.OrderResponse
	GetCurrentUserBalance(ctx context.Context) model.BalanceResponse
	WithdrawBalance(ctx context.Context, request model.WithdrawRequest) error
	GetWithdrawalsInfo(ctx context.Context) []model.Withdrawal
}
