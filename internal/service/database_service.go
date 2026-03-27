package service

import (
	"context"

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
	return ds.repo.RegisteUser(ctx, login, password)
}

func  (ds *DatabaseService) AuthenticateUser(ctx context.Context, login, password string) error {
	return ds.repo.AuthenticateUser(ctx, login, password)
}