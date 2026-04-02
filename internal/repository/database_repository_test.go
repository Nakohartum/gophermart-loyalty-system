package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/Nakohartum/gophermart-loyalty-system/internal/mocks"
	"github.com/Nakohartum/gophermart-loyalty-system/internal/model"
	"go.uber.org/mock/gomock"
)

func TestTextNewDatabaseRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	storage := mocks.NewMockStorage(ctrl)

	repository := NewDatabaseRepository(storage)
	if repository == nil {
		t.Fatal("NewDatabaseRepository returned nil")
	}
	if repository.database != storage {
		t.Fatal("NewDatabaseRepository did not store storage dependency")
	}
}

func TestTextDatabaseRepositoryDelegates(t *testing.T) {
	testCases := []struct {
		name string
		run  func(t *testing.T, storage *mocks.MockStorage, repository *DatabaseRepository)
	}{
		{
			name: "TextRegisterUser",
			run: func(t *testing.T, storage *mocks.MockStorage, repository *DatabaseRepository) {
				storage.EXPECT().RegisterUser(gomock.Any(), "user", "pass").Return(int64(42), nil)
				userID, err := repository.RegisterUser(context.Background(), "user", "pass")
				if err != nil || userID != 42 {
					t.Fatalf("RegisterUser() = (%d, %v), want (42, nil)", userID, err)
				}
			},
		},
		{
			name: "TextAuthenticateUser",
			run: func(t *testing.T, storage *mocks.MockStorage, repository *DatabaseRepository) {
				storage.EXPECT().AuthenticateUser(gomock.Any(), "user", "pass").Return(int64(42), nil)
				userID, err := repository.AuthenticateUser(context.Background(), "user", "pass")
				if err != nil || userID != 42 {
					t.Fatalf("AuthenticateUser() = (%d, %v), want (42, nil)", userID, err)
				}
			},
		},
		{
			name: "TextCreateOrder",
			run: func(t *testing.T, storage *mocks.MockStorage, repository *DatabaseRepository) {
				order := model.Order{Number: "123"}
				storage.EXPECT().CreateOrder(gomock.Any(), order, int64(7)).Return("123", nil)
				number, err := repository.CreateOrder(context.Background(), order, 7)
				if err != nil || number != "123" {
					t.Fatalf("CreateOrder() = (%q, %v), want (%q, nil)", number, err, "123")
				}
			},
		},
		{
			name: "TextGetOrdersForProcessing",
			run: func(t *testing.T, storage *mocks.MockStorage, repository *DatabaseRepository) {
				expected := []model.Order{{Number: "123"}}
				storage.EXPECT().GetOrdersForProcessing(gomock.Any()).Return(expected, nil)
				orders, err := repository.GetOrdersForProcessing(context.Background())
				if err != nil || len(orders) != 1 || orders[0].Number != "123" {
					t.Fatalf("GetOrdersForProcessing() = (%v, %v), want (%v, nil)", orders, err, expected)
				}
			},
		},
		{
			name: "TextUpdateOrder",
			run: func(t *testing.T, storage *mocks.MockStorage, repository *DatabaseRepository) {
				storage.EXPECT().UpdateOrder(gomock.Any(), int64(7), "123", model.PROCESSED, (*float64)(nil)).Return(nil)
				err := repository.UpdateOrder(context.Background(), 7, "123", model.PROCESSED, nil)
				if err != nil {
					t.Fatalf("UpdateOrder() error = %v, want nil", err)
				}
			},
		},
		{
			name: "TextGetListOfUploadedOrders",
			run: func(t *testing.T, storage *mocks.MockStorage, repository *DatabaseRepository) {
				expected := []model.OrderResponse{{Number: "123"}}
				storage.EXPECT().GetListOfUploadedOrders(gomock.Any(), int64(7)).Return(expected)
				orders := repository.GetListOfUploadedOrders(context.Background(), 7)
				if len(orders) != 1 || orders[0].Number != "123" {
					t.Fatalf("GetListOfUploadedOrders() = %v, want %v", orders, expected)
				}
			},
		},
		{
			name: "TextGetCurrentUserBalance",
			run: func(t *testing.T, storage *mocks.MockStorage, repository *DatabaseRepository) {
				expected := model.BalanceResponse{Current: 1, Withdrawn: 2}
				storage.EXPECT().GetCurrentUserBalance(gomock.Any(), int64(7)).Return(expected)
				balance := repository.GetCurrentUserBalance(context.Background(), 7)
				if balance != expected {
					t.Fatalf("GetCurrentUserBalance() = %+v, want %+v", balance, expected)
				}
			},
		},
		{
			name: "TextWithdrawBalance",
			run: func(t *testing.T, storage *mocks.MockStorage, repository *DatabaseRepository) {
				request := model.WithdrawRequest{UserID: 7, Order: "123", Sum: 10}
				storage.EXPECT().WithdrawBalance(gomock.Any(), request).Return(nil)
				err := repository.WithdrawBalance(context.Background(), request)
				if err != nil {
					t.Fatalf("WithdrawBalance() error = %v, want nil", err)
				}
			},
		},
		{
			name: "TextGetWithdrawalsInfo",
			run: func(t *testing.T, storage *mocks.MockStorage, repository *DatabaseRepository) {
				expected := []model.Withdrawal{{Order: "123", Sum: 10}}
				storage.EXPECT().GetWithdrawalsInfo(gomock.Any(), int64(7)).Return(expected)
				withdrawals := repository.GetWithdrawalsInfo(context.Background(), 7)
				if len(withdrawals) != 1 || withdrawals[0].Order != "123" {
					t.Fatalf("GetWithdrawalsInfo() = %v, want %v", withdrawals, expected)
				}
			},
		},
		{
			name: "TextPropagatesStorageError",
			run: func(t *testing.T, storage *mocks.MockStorage, repository *DatabaseRepository) {
				expectedErr := errors.New("storage failed")
				storage.EXPECT().RegisterUser(gomock.Any(), "user", "pass").Return(int64(0), expectedErr)
				_, err := repository.RegisterUser(context.Background(), "user", "pass")
				if !errors.Is(err, expectedErr) {
					t.Fatalf("RegisterUser() error = %v, want %v", err, expectedErr)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			storage := mocks.NewMockStorage(ctrl)
			repository := NewDatabaseRepository(storage)
			tc.run(t, storage, repository)
		})
	}
}
