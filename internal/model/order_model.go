package model

import "time"

type Status string

const (
	NEW        Status = "NEW"
	PROCESSING Status = "PROCESSING"
	INVALID    Status = "INVALID"
	PROCESSED  Status = "PROCESSED"
)

type Order struct {
	UserID     int64
	Number     string
	Status     Status
	Accrual    *float64
	UploadedAt time.Time
}

// DTO
type AccrualStatus string

const (
	Registered        AccrualStatus = "REGISTERED"
	AccrualInvalid    AccrualStatus = "INVALID"
	AccrualProcessing AccrualStatus = "PROCESSING"
	AccrualProcessed  AccrualStatus = "PROCESSED"
)

type AccrualOrder struct {
	Order   string        `json:"order"`
	Status  AccrualStatus `json:"status"`
	Accrual *float64      `json:"accrual,omitempty"`
}

type OrderResponse struct {
	Number     string    `json:"number"`
	Status     Status    `json:"status"`
	Accrual    *float64  `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}
