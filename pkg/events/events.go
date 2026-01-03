package events

import "time"

// Kafka topics
const (
	TopicOrderCreated = "order.created"
	TopicOrderDLQ     = "order.dlq"
)

// EventMetadata содержит общие метаданные для всех событий
type EventMetadata struct {
	EventID       string    `json:"event_id"`
	CorrelationID string    `json:"correlation_id"`
	Timestamp     time.Time `json:"timestamp"`
	Source        string    `json:"source"`
}

// OrderCreatedEvent публикуется при создании заказа
type OrderCreatedEvent struct {
	Metadata   EventMetadata `json:"metadata"`
	OrderID    string        `json:"order_id"`
	CustomerID string        `json:"customer_id"`
	Amount     float64       `json:"amount"`
	Currency   string        `json:"currency"`
}

// PaymentStatus представляет статус платежа
type PaymentStatus string

const (
	PaymentStatusPending PaymentStatus = "pending"
	PaymentStatusSuccess PaymentStatus = "success"
	PaymentStatusFailed  PaymentStatus = "failed"
)
