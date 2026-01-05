package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/yourusername/go-event-order-processor/services/order-api/internal/domain"
)

type OutboxRepository struct {
	db *sql.DB
}

func NewOutboxRepository(db *sql.DB) *OutboxRepository {
	return &OutboxRepository{db: db}
}

func (r *OutboxRepository) Create(ctx context.Context, msg *domain.OutboxMessage) error {
	query := `
		INSERT INTO outbox (aggregate_id, aggregate_type, event_type, payload, created_at, retry_count)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	err := r.db.QueryRowContext(ctx, query,
		msg.AggregateID,
		msg.AggregateType,
		msg.EventType,
		msg.Payload,
		msg.CreatedAt,
		msg.RetryCount,
	).Scan(&msg.ID)

	if err != nil {
		return fmt.Errorf("failed to create outbox message: %w", err)
	}

	return nil
}

func (r *OutboxRepository) GetUnprocessed(ctx context.Context, limit int) ([]*domain.OutboxMessage, error) {
	query := `
		SELECT id, aggregate_id, aggregate_type, event_type, payload, created_at, processed_at, retry_count, last_error
		FROM outbox
		WHERE processed_at IS NULL AND retry_count < 5
		ORDER BY created_at ASC
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query outbox: %w", err)
	}
	defer rows.Close()

	var messages []*domain.OutboxMessage
	for rows.Next() {
		var msg domain.OutboxMessage
		err := rows.Scan(
			&msg.ID,
			&msg.AggregateID,
			&msg.AggregateType,
			&msg.EventType,
			&msg.Payload,
			&msg.CreatedAt,
			&msg.ProcessedAt,
			&msg.RetryCount,
			&msg.LastError,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan outbox message: %w", err)
		}
		messages = append(messages, &msg)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return messages, nil
}

func (r *OutboxRepository) MarkProcessed(ctx context.Context, id int64) error {
	query := `
		UPDATE outbox
		SET processed_at = $1
		WHERE id = $2
	`

	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to mark outbox message as processed: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("outbox message not found: %d", id)
	}

	return nil
}

func (r *OutboxRepository) MarkFailed(ctx context.Context, id int64, failErr error) error {
	query := `
		UPDATE outbox
		SET retry_count = retry_count + 1,
		    last_error = $1
		WHERE id = $2
	`

	errMsg := failErr.Error()
	result, err := r.db.ExecContext(ctx, query, errMsg, id)
	if err != nil {
		return fmt.Errorf("failed to mark outbox message as failed: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("outbox message not found: %d", id)
	}

	return nil
}
