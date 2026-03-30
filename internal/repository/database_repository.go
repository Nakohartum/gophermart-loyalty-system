package repository

import (
	"context"

	"github.com/Nakohartum/gophermart-loyalty-system/internal/config/db"
	"github.com/Nakohartum/gophermart-loyalty-system/internal/model"
)

type DatabaseRepository struct {
	database *db.PgDatabase
}

func NewDatabaseRepository(database *db.PgDatabase) *DatabaseRepository{
	return &DatabaseRepository{
		database: database,
	}
}

func (dr *DatabaseRepository) RegisterUser(ctx context.Context, username, password string) (int64, error){
	return dr.database.RegisterUser(ctx, username, password)
}

func (dr *DatabaseRepository) AuthenticateUser(ctx context.Context, username, password string) (int64, error) {
	return dr.database.AuthenticateUser(ctx, username, password)
}

func (dr *DatabaseRepository) CreateOrder(ctx context.Context, order model.Order, userId int64) (string, error){
	return dr.database.CreateOrder(ctx, order, userId)
}

func (dr *DatabaseRepository) GetOrdersForProcessing(ctx context.Context) ([]model.Order, error) {
	return dr.database.GetOrdersForProcessing(ctx)
}

func (dr *DatabaseRepository) UpdateOrder(ctx context.Context, userId int64, number string, status model.Status, accrual *float64) error {
	return dr.database.UpdateOrder(ctx, userId, number, status, accrual)
}

func (dr *DatabaseRepository) GetListOfUploadedOrders(ctx context.Context, userId int64) []model.OrderResponse {
	return dr.database.GetListOfUploadedOrders(ctx, userId)
}