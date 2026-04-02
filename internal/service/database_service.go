package service

import (
	"context"

	"github.com/Nakohartum/gophermart-loyalty-system/internal/model"
	"github.com/Nakohartum/gophermart-loyalty-system/internal/repository"
)

type DatabaseService struct {
	repo repository.Repository
}

func NewDatabaseRepo(repo repository.Repository) *DatabaseService {
	return &DatabaseService{
		repo: repo,
	}
}

func (ds *DatabaseService) RegisterUser(ctx context.Context, login, password string) (int64, error) {
	return ds.repo.RegisterUser(ctx, login, password)
}

func (ds *DatabaseService) AuthenticateUser(ctx context.Context, login, password string) (int64, error) {
	return ds.repo.AuthenticateUser(ctx, login, password)
}

func (ds *DatabaseService) CreateOrder(ctx context.Context, order model.Order, userID int64) (string, error) {
	return ds.repo.CreateOrder(ctx, order, userID)
}

func (ds *DatabaseService) GetOrdersForProcessing(ctx context.Context) ([]model.Order, error) {
	return ds.repo.GetOrdersForProcessing(ctx)
}

func (ds *DatabaseService) UpdateOrder(ctx context.Context, userID int64, number string, status model.Status, accrual *float64) error {
	return ds.repo.UpdateOrder(ctx, userID, number, status, accrual)
}

func (ds *DatabaseService) GetListOfUploadedOrders(ctx context.Context, userID int64) []model.OrderResponse {
	return ds.repo.GetListOfUploadedOrders(ctx, userID)
}

func (ds *DatabaseService) GetCurrentUserBalance(ctx context.Context, userID int64) model.BalanceResponse {
	return ds.repo.GetCurrentUserBalance(ctx, userID)
}

func (ds *DatabaseService) WithdrawBalance(ctx context.Context, request model.WithdrawRequest) error {
	return ds.repo.WithdrawBalance(ctx, request)
}

func (ds *DatabaseService) GetWithdrawalsInfo(ctx context.Context, userID int64) []model.Withdrawal {
	return ds.repo.GetWithdrawalsInfo(ctx, userID)
}
