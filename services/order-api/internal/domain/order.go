package domain

import (
	"errors"
	"time"
)

var (
	ErrInvalidAmount     = errors.New("invalid amount")
	ErrInvalidCustomerID = errors.New("invalid customer ID")
	ErrOrderNotFound     = errors.New("order not found")
)

type OrderStatus string

const (
	StatusPending   OrderStatus = "pending"
	StatusPaid      OrderStatus = "paid"
	StatusFailed    OrderStatus = "failed"
	StatusCancelled OrderStatus = "cancelled"
)

type Order struct {
	ID         string
	CustomerID string
	Amount     float64
	Currency   string
	Status     OrderStatus
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func NewOrder(customerID string, amount float64) (*Order, error) {
	if customerID == "" {
		return nil, ErrInvalidCustomerID
	}
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}

	now := time.Now()
	return &Order{
		CustomerID: customerID,
		Amount:     amount,
		Currency:   "USD",
		Status:     StatusPending,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

func (o *Order) UpdateStatus(status OrderStatus) {
	o.Status = status
	o.UpdatedAt = time.Now()
}
