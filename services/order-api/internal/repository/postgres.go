package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/yourusername/go-event-order-processor/services/order-api/internal/domain"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// CreateWithOutbox creates order and outbox message in a single transaction
func (r *PostgresRepository) CreateWithOutbox(ctx context.Context, order *domain.Order, outboxMsg *domain.OutboxMessage) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert order
	orderQuery := `
		INSERT INTO orders (id, customer_id, amount, currency, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err = tx.ExecContext(ctx, orderQuery,
		order.ID,
		order.CustomerID,
		order.Amount,
		order.Currency,
		order.Status,
		order.CreatedAt,
		order.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}

	// Insert outbox message
	outboxQuery := `
		INSERT INTO outbox (aggregate_id, aggregate_type, event_type, payload, created_at, retry_count)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	err = tx.QueryRowContext(ctx, outboxQuery,
		outboxMsg.AggregateID,
		outboxMsg.AggregateType,
		outboxMsg.EventType,
		outboxMsg.Payload,
		outboxMsg.CreatedAt,
		outboxMsg.RetryCount,
	).Scan(&outboxMsg.ID)
	if err != nil {
		return fmt.Errorf("failed to create outbox message: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *PostgresRepository) Create(ctx context.Context, order *domain.Order) error {
	query := `
		INSERT INTO orders (id, customer_id, amount, currency, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.ExecContext(ctx, query,
		order.ID,
		order.CustomerID,
		order.Amount,
		order.Currency,
		order.Status,
		order.CreatedAt,
		order.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}

	return nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	query := `
		SELECT id, customer_id, amount, currency, status, created_at, updated_at
		FROM orders
		WHERE id = $1
	`

	var order domain.Order
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&order.ID,
		&order.CustomerID,
		&order.Amount,
		&order.Currency,
		&order.Status,
		&order.CreatedAt,
		&order.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrOrderNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	return &order, nil
}

func (r *PostgresRepository) UpdateStatus(ctx context.Context, id string, status domain.OrderStatus) error {
	query := `
		UPDATE orders
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`

	result, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return domain.ErrOrderNotFound
	}

	return nil
}

func NewDB(connString string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)
	db.SetConnMaxIdleTime(10 * time.Minute)

	return db, nil
}
