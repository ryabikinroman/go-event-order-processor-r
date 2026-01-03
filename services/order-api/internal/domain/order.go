package domain

import (
	"errors"
	"time"
)

var (
	ErrInvalidAmount     = errors.New("invalid amount")
	ErrInvalidCustomerID = errors.New("invalid customer ID")
	ErrInvalidCurrency   = errors.New("invalid currency")
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

var validCurrencies = map[string]bool{
	"USD": true,
	"EUR": true,
	"GBP": true,
	"JPY": true,
	"CNY": true,
}

func NewOrder(customerID string, amount float64) (*Order, error) {
	return NewOrderWithCurrency(customerID, amount, "USD")
}

func NewOrderWithCurrency(customerID string, amount float64, currency string) (*Order, error) {
	if customerID == "" {
		return nil, ErrInvalidCustomerID
	}
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}
	if !validCurrencies[currency] {
		return nil, ErrInvalidCurrency
	}

	now := time.Now()
	return &Order{
		CustomerID: customerID,
		Amount:     amount,
		Currency:   currency,
		Status:     StatusPending,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

func (o *Order) UpdateStatus(status OrderStatus) {
	o.Status = status
	o.UpdatedAt = time.Now()
}
