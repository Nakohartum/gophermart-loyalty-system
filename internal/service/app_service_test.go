package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Nakohartum/gophermart-loyalty-system/internal/apperrors"
	"github.com/Nakohartum/gophermart-loyalty-system/internal/mocks"
	"github.com/Nakohartum/gophermart-loyalty-system/internal/model"
	"go.uber.org/mock/gomock"
)

func TestTextNewAppService(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockRepository(ctrl)
	auth := mocks.NewMockAuth(ctrl)

	service := NewAppService(repo, auth)
	if service == nil {
		t.Fatal("NewAppService returned nil")
	}
	if service.repo != repo {
		t.Fatal("NewAppService did not store repository dependency")
	}
	if service.auth != auth {
		t.Fatal("NewAppService did not store auth dependency")
	}
}

func TestTextRegisterUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockRepository(ctrl)
	auth := mocks.NewMockAuth(ctrl)
	service := NewAppService(repo, auth)

	repo.EXPECT().RegisterUser(gomock.Any(), "user", "pass").Return(int64(42), nil)
	auth.EXPECT().GenerateToken(int64(42)).Return("token", nil)

	token, err := service.RegisterUser(context.Background(), "user", "pass")
	if err != nil || token != "token" {
		t.Fatalf("RegisterUser() = (%q, %v), want (%q, nil)", token, err, "token")
	}
}

func TestTextAuthenticateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockRepository(ctrl)
	auth := mocks.NewMockAuth(ctrl)
	service := NewAppService(repo, auth)

	repo.EXPECT().AuthenticateUser(gomock.Any(), "user", "pass").Return(int64(42), nil)
	auth.EXPECT().GenerateToken(int64(42)).Return("token", nil)

	token, err := service.AuthenticateUser(context.Background(), "user", "pass")
	if err != nil || token != "token" {
		t.Fatalf("AuthenticateUser() = (%q, %v), want (%q, nil)", token, err, "token")
	}
}

func TestTextCreateOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockRepository(ctrl)
	auth := mocks.NewMockAuth(ctrl)
	service := NewAppService(repo, auth)

	repo.EXPECT().CreateOrder(gomock.Any(), gomock.AssignableToTypeOf(model.Order{}), int64(7)).DoAndReturn(
		func(ctx context.Context, order model.Order, userID int64) (string, error) {
			if order.Number != "79927398713" || order.Status != model.NEW || userID != 7 {
				t.Fatalf("unexpected order payload: %#v userID=%d", order, userID)
			}
			if time.Since(order.UploadedAt) > time.Second {
				t.Fatalf("UploadedAt is too old: %v", order.UploadedAt)
			}
			return order.Number, nil
		},
	)

	err := service.CreateOrder(context.Background(), 7, " 79927398713 ")
	if err != nil {
		t.Fatalf("CreateOrder() error = %v, want nil", err)
	}
}

func TestTextCreateOrderInvalidNumber(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockRepository(ctrl)
	auth := mocks.NewMockAuth(ctrl)
	service := NewAppService(repo, auth)

	err := service.CreateOrder(context.Background(), 7, "123")
	if !errors.Is(err, apperrors.ErrInvalidOrderNumber) {
		t.Fatalf("CreateOrder() error = %v, want %v", err, apperrors.ErrInvalidOrderNumber)
	}
}

func TestTextAppServiceDelegates(t *testing.T) {
	testCases := []struct {
		name string
		run  func(t *testing.T, repo *mocks.MockRepository, service *AppService)
	}{
		{
			name: "TextGetOrdersForProcessing",
			run: func(t *testing.T, repo *mocks.MockRepository, service *AppService) {
				expected := []model.Order{{Number: "123"}}
				repo.EXPECT().GetOrdersForProcessing(gomock.Any()).Return(expected, nil)
				orders, err := service.GetOrdersForProcessing(context.Background())
				if err != nil || len(orders) != 1 || orders[0].Number != "123" {
					t.Fatalf("GetOrdersForProcessing() = (%v, %v), want (%v, nil)", orders, err, expected)
				}
			},
		},
		{
			name: "TextUpdateOrder",
			run: func(t *testing.T, repo *mocks.MockRepository, service *AppService) {
				repo.EXPECT().UpdateOrder(gomock.Any(), int64(7), "123", model.PROCESSED, (*float64)(nil)).Return(nil)
				err := service.UpdateOrder(context.Background(), 7, "123", model.PROCESSED, nil)
				if err != nil {
					t.Fatalf("UpdateOrder() error = %v, want nil", err)
				}
			},
		},
		{
			name: "TextGetListOfUploadedOrders",
			run: func(t *testing.T, repo *mocks.MockRepository, service *AppService) {
				expected := []model.OrderResponse{{Number: "123"}}
				repo.EXPECT().GetListOfUploadedOrders(gomock.Any(), int64(7)).Return(expected)
				orders := service.GetListOfUploadedOrders(context.Background(), 7)
				if len(orders) != 1 || orders[0].Number != "123" {
					t.Fatalf("GetListOfUploadedOrders() = %v, want %v", orders, expected)
				}
			},
		},
		{
			name: "TextGetCurrentUserBalance",
			run: func(t *testing.T, repo *mocks.MockRepository, service *AppService) {
				expected := model.BalanceResponse{Current: 1, Withdrawn: 2}
				repo.EXPECT().GetCurrentUserBalance(gomock.Any(), int64(7)).Return(expected)
				balance := service.GetCurrentUserBalance(context.Background(), 7)
				if balance != expected {
					t.Fatalf("GetCurrentUserBalance() = %+v, want %+v", balance, expected)
				}
			},
		},
		{
			name: "TextWithdrawBalance",
			run: func(t *testing.T, repo *mocks.MockRepository, service *AppService) {
				request := model.WithdrawRequest{UserID: 7, Order: "123", Sum: 10}
				repo.EXPECT().WithdrawBalance(gomock.Any(), request).Return(nil)
				err := service.WithdrawBalance(context.Background(), request)
				if err != nil {
					t.Fatalf("WithdrawBalance() error = %v, want nil", err)
				}
			},
		},
		{
			name: "TextGetWithdrawalsInfo",
			run: func(t *testing.T, repo *mocks.MockRepository, service *AppService) {
				expected := []model.Withdrawal{{Order: "123", Sum: 10}}
				repo.EXPECT().GetWithdrawalsInfo(gomock.Any(), int64(7)).Return(expected)
				withdrawals := service.GetWithdrawalsInfo(context.Background(), 7)
				if len(withdrawals) != 1 || withdrawals[0].Order != "123" {
					t.Fatalf("GetWithdrawalsInfo() = %v, want %v", withdrawals, expected)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := mocks.NewMockRepository(ctrl)
			auth := mocks.NewMockAuth(ctrl)
			service := NewAppService(repo, auth)
			tc.run(t, repo, service)
		})
	}
}
