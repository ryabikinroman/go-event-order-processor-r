package idempotency

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Checker struct {
	client *redis.Client
	ttl    time.Duration
}

func NewChecker(client *redis.Client, ttl time.Duration) *Checker {
	return &Checker{
		client: client,
		ttl:    ttl,
	}
}

// CheckAndSet проверяет, было ли событие уже обработано
// Возвращает true, если это первая обработка (нужно продолжить)
// Возвращает false, если уже было обработано (пропустить)
func (c *Checker) CheckAndSet(ctx context.Context, eventID string) (bool, error) {
	key := fmt.Sprintf("processed:%s", eventID)

	// SetNX возвращает true, если ключ был установлен (первый раз)
	wasSet, err := c.client.SetNX(ctx, key, "1", c.ttl).Result()
	if err != nil {
		return false, fmt.Errorf("redis setnx failed: %w", err)
	}

	return wasSet, nil
}

// Remove удаляет маркер идемпотентности (для повторной обработки)
func (c *Checker) Remove(ctx context.Context, eventID string) error {
	key := fmt.Sprintf("processed:%s", eventID)
	return c.client.Del(ctx, key).Err()
}
