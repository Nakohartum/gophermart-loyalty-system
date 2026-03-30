package service

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/Nakohartum/gophermart-loyalty-system/internal/model"
)

type OrderProcessor struct {
	databaseService *DatabaseService
	accrualService  *AccrualService
	interval        time.Duration
}

func NewOrderProcessor(databaseService *DatabaseService, accrualService *AccrualService, interval time.Duration) *OrderProcessor {
	return &OrderProcessor{
		databaseService: databaseService,
		accrualService:  accrualService,
		interval:        interval,
	}
}

func (op *OrderProcessor) Start(ctx context.Context) {
	ticker := time.NewTicker(op.interval)
	defer ticker.Stop()

	op.processPendingOrders(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			op.processPendingOrders(ctx)
		}
	}
}

func (op *OrderProcessor) processPendingOrders(ctx context.Context) {
	orders, err := op.databaseService.GetOrdersForProcessing(ctx)
	if err != nil {
		log.Printf("load orders for processing: %v", err)
		return
	}

	for _, order := range orders {
		if err := op.processOrder(ctx, order); err != nil {
			log.Printf("process order %s: %v", order.Number, err)
		}
		log.Printf("process order %s done", order.Number)
	}
}

func (op *OrderProcessor) processOrder(ctx context.Context, order model.Order) error {
	accrualOrder, err := op.accrualService.GetOrder(ctx, order.Number)
	if err != nil {
		if errors.Is(err, ErrAccrualTryLater) {
			log.Println(err.Error())
			return nil
		}
		return err
	}
	if accrualOrder == nil {
		log.Println("no order")
		return nil
	}

	status := mapAccrualStatus(accrualOrder.Status)
	return op.databaseService.UpdateOrder(ctx, order.UserID, order.Number, status, accrualOrder.Accrual)
}

func mapAccrualStatus(status model.AccrualStatus) model.Status {
	switch status {
	case model.ACCRUAL_INVALID:
		return model.INVALID
	case model.ACCRUAL_PROCESSED:
		return model.PROCESSED
	case model.ACCRUAL_PROCESSING, model.REGISTERED:
		return model.PROCESSING
	default:
		return model.PROCESSING
	}
}
