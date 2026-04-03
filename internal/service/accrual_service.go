package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Nakohartum/gophermart-loyalty-system/internal/model"
)

var ErrAccrualTryLater = errors.New("accrual asked to retry later")

type AccrualService struct {
	baseURL string
	client  *http.Client
}

func NewAccrualService(baseURL string) *AccrualService {
	return &AccrualService{
		baseURL: strings.TrimRight(baseURL, "/"),
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (as *AccrualService) GetOrder(ctx context.Context, number string) (*model.AccrualOrder, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, as.baseURL+"/api/orders/"+number, nil)
	if err != nil {
		return nil, err
	}

	resp, err := as.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var order model.AccrualOrder
		if err := json.NewDecoder(resp.Body).Decode(&order); err != nil {
			return nil, err
		}
		return &order, nil
	case http.StatusNoContent:
		log.Println("Status no content")
		return nil, nil
	case http.StatusTooManyRequests:
		return nil, ErrAccrualTryLater
	default:
		return nil, fmt.Errorf("unexpected accrual status: %d", resp.StatusCode)
	}
}
