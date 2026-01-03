package domain

import (
	"context"
	"time"
)

const (
	AggregateTypeOrder = "order"
	EventTypeOrderCreated = "order.created"
)

type OutboxMessage struct {
	ID            int64
	AggregateID   string
	AggregateType string
	EventType     string
	Payload       []byte
	CreatedAt     time.Time
	ProcessedAt   *time.Time
	RetryCount    int
	LastError     *string
}

type OutboxRepository interface {
	Create(ctx context.Context, msg *OutboxMessage) error
	GetUnprocessed(ctx context.Context, limit int) ([]*OutboxMessage, error)
	MarkProcessed(ctx context.Context, id int64) error
	MarkFailed(ctx context.Context, id int64, err error) error
}
