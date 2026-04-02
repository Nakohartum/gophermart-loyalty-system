package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Nakohartum/gophermart-loyalty-system/internal/mocks"
	"github.com/Nakohartum/gophermart-loyalty-system/internal/model"
	"go.uber.org/mock/gomock"
)

func TestTextMapAccrualStatus(t *testing.T) {
	testCases := []struct {
		name     string
		input    model.AccrualStatus
		expected model.Status
	}{
		{name: "invalid", input: model.AccrualInvalid, expected: model.INVALID},
		{name: "processed", input: model.AccrualProcessed, expected: model.PROCESSED},
		{name: "processing", input: model.AccrualProcessing, expected: model.PROCESSING},
		{name: "registered", input: model.Registered, expected: model.PROCESSING},
		{name: "unknown", input: model.AccrualStatus("OTHER"), expected: model.PROCESSING},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := mapAccrualStatus(tc.input); got != tc.expected {
				t.Fatalf("mapAccrualStatus(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}

func TestTextProcessOrder(t *testing.T) {
	testCases := []struct {
		name      string
		setup     func(database *mocks.MockDatabase, accrual *mocks.MockAccrual)
		expectErr error
	}{
		{
			name: "try later returns nil",
			setup: func(database *mocks.MockDatabase, accrual *mocks.MockAccrual) {
				accrual.EXPECT().GetOrder(gomock.Any(), "123").Return(nil, ErrAccrualTryLater)
			},
		},
		{
			name: "nil order returns nil",
			setup: func(database *mocks.MockDatabase, accrual *mocks.MockAccrual) {
				accrual.EXPECT().GetOrder(gomock.Any(), "123").Return(nil, nil)
			},
		},
		{
			name: "accrual error returns error",
			setup: func(database *mocks.MockDatabase, accrual *mocks.MockAccrual) {
				accrual.EXPECT().GetOrder(gomock.Any(), "123").Return(nil, errors.New("accrual failed"))
			},
			expectErr: errors.New("accrual failed"),
		},
		{
			name: "update order success",
			setup: func(database *mocks.MockDatabase, accrual *mocks.MockAccrual) {
				value := 10.5
				accrual.EXPECT().GetOrder(gomock.Any(), "123").Return(&model.AccrualOrder{
					Order:   "123",
					Status:  model.AccrualProcessed,
					Accrual: &value,
				}, nil)
				database.EXPECT().UpdateOrder(gomock.Any(), int64(77), "123", model.PROCESSED, &value).Return(nil)
			},
		},
		{
			name: "update order error",
			setup: func(database *mocks.MockDatabase, accrual *mocks.MockAccrual) {
				accrual.EXPECT().GetOrder(gomock.Any(), "123").Return(&model.AccrualOrder{
					Order:  "123",
					Status: model.AccrualInvalid,
				}, nil)
				database.EXPECT().UpdateOrder(gomock.Any(), int64(77), "123", model.INVALID, (*float64)(nil)).Return(errors.New("update failed"))
			},
			expectErr: errors.New("update failed"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			database := mocks.NewMockDatabase(ctrl)
			accrual := mocks.NewMockAccrual(ctrl)
			tc.setup(database, accrual)

			processor := NewOrderProcessor(database, accrual, time.Millisecond)
			err := processor.processOrder(context.Background(), model.Order{UserID: 77, Number: "123"})

			if tc.expectErr != nil {
				if err == nil || err.Error() != tc.expectErr.Error() {
					t.Fatalf("processOrder error = %v, want %v", err, tc.expectErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("processOrder returned error: %v", err)
			}
		})
	}
}

func TestTextProcessPendingOrders(t *testing.T) {
	testCases := []struct {
		name  string
		setup func(database *mocks.MockDatabase, accrual *mocks.MockAccrual)
	}{
		{
			name: "load orders error",
			setup: func(database *mocks.MockDatabase, accrual *mocks.MockAccrual) {
				database.EXPECT().GetOrdersForProcessing(gomock.Any()).Return(nil, errors.New("load failed"))
			},
		},
		{
			name: "processes loaded orders",
			setup: func(database *mocks.MockDatabase, accrual *mocks.MockAccrual) {
				database.EXPECT().GetOrdersForProcessing(gomock.Any()).Return([]model.Order{
					{UserID: 1, Number: "111"},
					{UserID: 2, Number: "222"},
				}, nil)
				accrual.EXPECT().GetOrder(gomock.Any(), "111").Return(nil, nil)
				accrual.EXPECT().GetOrder(gomock.Any(), "222").Return(nil, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			database := mocks.NewMockDatabase(ctrl)
			accrual := mocks.NewMockAccrual(ctrl)
			tc.setup(database, accrual)

			processor := NewOrderProcessor(database, accrual, time.Millisecond)
			processor.processPendingOrders(context.Background())
		})
	}
}

func TestTextStart(t *testing.T) {
	ctrl := gomock.NewController(t)
	database := mocks.NewMockDatabase(ctrl)
	accrual := mocks.NewMockAccrual(ctrl)
	processor := NewOrderProcessor(database, accrual, time.Hour)

	ctx, cancel := context.WithCancel(context.Background())
	database.EXPECT().GetOrdersForProcessing(gomock.Any()).DoAndReturn(func(context.Context) ([]model.Order, error) {
		cancel()
		return nil, nil
	})

	done := make(chan struct{})
	go func() {
		defer close(done)
		processor.Start(ctx)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Start did not stop after context cancellation")
	}
}
