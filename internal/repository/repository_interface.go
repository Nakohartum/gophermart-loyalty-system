package repository

import (
	"context"

	"github.com/Nakohartum/gophermart-loyalty-system/internal/model"
)

type Repository interface {
	RegisterUser(ctx context.Context, login, password string) (int64, error)
	AuthenticateUser(ctx context.Context, login, password string) (int64, error)
	CreateOrder(ctx context.Context, order model.Order, userId int64) (string, error)
	GetOrdersForProcessing(ctx context.Context) ([]model.Order, error)
	GetListOfUploadedOrders(ctx context.Context) []model.OrderResponse
	GetCurrentUserBalance(ctx context.Context) model.BalanceResponse
	WithdrawBalance(ctx context.Context, request model.WithdrawRequest) error
	GetWithdrawalsInfo(ctx context.Context) []model.Withdrawal
}
