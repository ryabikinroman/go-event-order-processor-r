package producer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"github.com/yourusername/go-event-order-processor/pkg/events"
	"github.com/yourusername/go-event-order-processor/pkg/logger"
	"github.com/yourusername/go-event-order-processor/services/order-api/internal/domain"
)

type KafkaProducer struct {
	producer sarama.SyncProducer
	log      *logger.Logger
}

func NewKafkaProducer(brokers []string, log *logger.Logger) (*KafkaProducer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 3
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka producer: %w", err)
	}

	return &KafkaProducer{
		producer: producer,
		log:      log,
	}, nil
}

func (p *KafkaProducer) PublishOrderCreated(ctx context.Context, order *domain.Order) error {
	// Get correlation ID from context (set by logger package)
	correlationID := getCorrelationID(ctx)
	if correlationID == "" {
		correlationID = uuid.New().String()
	}

	event := events.OrderCreatedEvent{
		Metadata: events.EventMetadata{
			EventID:       uuid.New().String(),
			CorrelationID: correlationID,
			Timestamp:     time.Now(),
			Source:        "order-api",
		},
		OrderID:    order.ID,
		CustomerID: order.CustomerID,
		Amount:     order.Amount,
		Currency:   order.Currency,
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: events.TopicOrderCreated,
		Key:   sarama.StringEncoder(order.ID),
		Value: sarama.ByteEncoder(payload),
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		p.log.Errorw("Failed to send message", "error", err)
		return fmt.Errorf("failed to send message: %w", err)
	}

	p.log.Infow("Event published",
		"event_id", event.Metadata.EventID,
		"correlation_id", correlationID,
		"order_id", order.ID,
		"partition", partition,
		"offset", offset,
	)

	return nil
}

func (p *KafkaProducer) Close() error {
	return p.producer.Close()
}

// getCorrelationID extracts correlation ID from context
// Uses the same context key as logger package
func getCorrelationID(ctx context.Context) string {
	// This must match the key used in logger.WithCorrelationID
	type contextKey string
	const correlationIDKey contextKey = "correlation_id"

	if id, ok := ctx.Value(correlationIDKey).(string); ok {
		return id
	}
	return ""
}
