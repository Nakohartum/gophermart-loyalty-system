package service

import (
	"context"
	"strings"
	"time"

	"github.com/Nakohartum/gophermart-loyalty-system/internal/apperrors"
	"github.com/Nakohartum/gophermart-loyalty-system/internal/model"
	"github.com/Nakohartum/gophermart-loyalty-system/internal/repository"
)

type AppService struct {
	repo repository.Repository
	auth Auth
}

func NewAppService(repo repository.Repository, auth Auth) *AppService {
	return &AppService{
		repo: repo,
		auth: auth,
	}
}

func (s *AppService) RegisterUser(ctx context.Context, login, password string) (string, error) {
	userID, err := s.repo.RegisterUser(ctx, login, password)
	if err != nil {
		return "", err
	}

	return s.auth.GenerateToken(userID)
}

func (s *AppService) AuthenticateUser(ctx context.Context, login, password string) (string, error) {
	userID, err := s.repo.AuthenticateUser(ctx, login, password)
	if err != nil {
		return "", err
	}

	return s.auth.GenerateToken(userID)
}

func (s *AppService) CreateOrder(ctx context.Context, userID int64, orderNumber string) error {
	orderNumber = strings.TrimSpace(orderNumber)
	if !IsValidLuhn(orderNumber) {
		return apperrors.ErrInvalidOrderNumber
	}

	_, err := s.repo.CreateOrder(ctx, model.Order{
		Number:     orderNumber,
		Status:     model.NEW,
		UploadedAt: time.Now(),
	}, userID)
	return err
}

func (s *AppService) GetListOfUploadedOrders(ctx context.Context, userID int64) []model.OrderResponse {
	return s.repo.GetListOfUploadedOrders(ctx, userID)
}

func (s *AppService) GetCurrentUserBalance(ctx context.Context, userID int64) model.BalanceResponse {
	return s.repo.GetCurrentUserBalance(ctx, userID)
}

func (s *AppService) WithdrawBalance(ctx context.Context, request model.WithdrawRequest) error {
	return s.repo.WithdrawBalance(ctx, request)
}

func (s *AppService) GetWithdrawalsInfo(ctx context.Context, userID int64) []model.Withdrawal {
	return s.repo.GetWithdrawalsInfo(ctx, userID)
}

func (s *AppService) GetOrdersForProcessing(ctx context.Context) ([]model.Order, error) {
	return s.repo.GetOrdersForProcessing(ctx)
}

func (s *AppService) UpdateOrder(ctx context.Context, userID int64, number string, status model.Status, accrual *float64) error {
	return s.repo.UpdateOrder(ctx, userID, number, status, accrual)
}
