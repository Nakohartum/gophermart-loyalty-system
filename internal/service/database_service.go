package service

import (
	"context"

	"github.com/Nakohartum/gophermart-loyalty-system/internal/model"
	"github.com/Nakohartum/gophermart-loyalty-system/internal/repository"
)

type DatabaseService struct {
	repo *repository.DatabaseRepository
}

func NewDatabaseRepo(repo *repository.DatabaseRepository) *DatabaseService {
	return &DatabaseService{
		repo: repo,
	}
}

func  (ds *DatabaseService) RegisterUser(ctx context.Context, login, password string) (int64, error) {
	return ds.repo.RegisterUser(ctx, login, password)
}

func  (ds *DatabaseService) AuthenticateUser(ctx context.Context, login, password string) (int64, error) {
	return ds.repo.AuthenticateUser(ctx, login, password)
}

func (ds *DatabaseService) CreateOrder(ctx context.Context, order model.Order, userId int64) (string, error){
	return ds.repo.CreateOrder(ctx, order, userId)
}

func (ds *DatabaseService) GetOrdersForProcessing(ctx context.Context) ([]model.Order, error) {
	return ds.repo.GetOrdersForProcessing(ctx)
}

func (ds *DatabaseService) UpdateOrder(ctx context.Context, userId int64, number string, status model.Status, accrual *float64) error {
	return ds.repo.UpdateOrder(ctx, userId, number, status, accrual)
}

func (ds *DatabaseService) GetListOfUploadedOrders(ctx context.Context, userId int64) []model.OrderResponse {
	return ds.repo.GetListOfUploadedOrders(ctx, userId)
}
