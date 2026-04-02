package repository

import (
	"context"

	"github.com/Nakohartum/gophermart-loyalty-system/internal/config/db"
	"github.com/Nakohartum/gophermart-loyalty-system/internal/model"
)

type DatabaseRepository struct {
	database db.Storage
}

func NewDatabaseRepository(database db.Storage) *DatabaseRepository {
	return &DatabaseRepository{
		database: database,
	}
}

func (dr *DatabaseRepository) RegisterUser(ctx context.Context, username, password string) (int64, error) {
	return dr.database.RegisterUser(ctx, username, password)
}

func (dr *DatabaseRepository) AuthenticateUser(ctx context.Context, username, password string) (int64, error) {
	return dr.database.AuthenticateUser(ctx, username, password)
}

func (dr *DatabaseRepository) CreateOrder(ctx context.Context, order model.Order, userID int64) (string, error) {
	return dr.database.CreateOrder(ctx, order, userID)
}

func (dr *DatabaseRepository) GetOrdersForProcessing(ctx context.Context) ([]model.Order, error) {
	return dr.database.GetOrdersForProcessing(ctx)
}

func (dr *DatabaseRepository) UpdateOrder(ctx context.Context, userID int64, number string, status model.Status, accrual *float64) error {
	return dr.database.UpdateOrder(ctx, userID, number, status, accrual)
}

func (dr *DatabaseRepository) GetListOfUploadedOrders(ctx context.Context, userID int64) []model.OrderResponse {
	return dr.database.GetListOfUploadedOrders(ctx, userID)
}

func (dr *DatabaseRepository) GetCurrentUserBalance(ctx context.Context, userID int64) model.BalanceResponse {
	return dr.database.GetCurrentUserBalance(ctx, userID)
}

func (dr *DatabaseRepository) WithdrawBalance(ctx context.Context, request model.WithdrawRequest) error {
	return dr.database.WithdrawBalance(ctx, request)
}

func (dr *DatabaseRepository) GetWithdrawalsInfo(ctx context.Context, userID int64) []model.Withdrawal {
	return dr.database.GetWithdrawalsInfo(ctx, userID)
}
