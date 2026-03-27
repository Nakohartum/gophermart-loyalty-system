package repository

import (
	"context"

	"github.com/Nakohartum/gophermart-loyalty-system/internal/config/db"
)

type DatabaseRepository struct {
	database *db.PgDatabase
}

func NewDatabaseRepository(database *db.PgDatabase) *DatabaseRepository{
	return &DatabaseRepository{
		database: database,
	}
}

func (dr *DatabaseRepository) RegisteUser(ctx context.Context, username, password string) (int64, error){
	return dr.database.RegisterUser(ctx, username, password)
}

func (dr *DatabaseRepository) AuthenticateUser(ctx context.Context, username, password string) error {
	return dr.database.AuthenticateUser(ctx, username, password)
}