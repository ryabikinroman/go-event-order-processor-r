package outbox

import (
	"context"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/yourusername/go-event-order-processor/pkg/events"
	"github.com/yourusername/go-event-order-processor/pkg/logger"
	"github.com/yourusername/go-event-order-processor/services/order-api/internal/domain"
)

var (
	outboxProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "outbox_messages_processed_total",
		Help: "Total number of outbox messages successfully processed",
	})

	outboxFailed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "outbox_messages_failed_total",
		Help: "Total number of outbox messages that failed to process",
	})

	outboxPending = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "outbox_messages_pending",
		Help: "Current number of pending outbox messages",
	})
)

type Processor struct {
	repo     domain.OutboxRepository
	producer sarama.SyncProducer
	log      *logger.Logger
	interval time.Duration
	batchSize int
}

func NewProcessor(
	repo domain.OutboxRepository,
	producer sarama.SyncProducer,
	log *logger.Logger,
	interval time.Duration,
	batchSize int,
) *Processor {
	return &Processor{
		repo:      repo,
		producer:  producer,
		log:       log,
		interval:  interval,
		batchSize: batchSize,
	}
}

func (p *Processor) Start(ctx context.Context) error {
	p.log.Info("Outbox processor started")
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	// Process immediately on start
	p.processBatch(ctx)

	for {
		select {
		case <-ctx.Done():
			p.log.Info("Outbox processor stopped")
			return ctx.Err()
		case <-ticker.C:
			p.processBatch(ctx)
		}
	}
}

func (p *Processor) processBatch(ctx context.Context) {
	messages, err := p.repo.GetUnprocessed(ctx, p.batchSize)
	if err != nil {
		p.log.Errorw("Failed to get unprocessed outbox messages", "error", err)
		return
	}

	if len(messages) == 0 {
		outboxPending.Set(0)
		return
	}

	p.log.Infow("Processing outbox batch", "count", len(messages))
	outboxPending.Set(float64(len(messages)))

	for _, msg := range messages {
		if err := p.processMessage(ctx, msg); err != nil {
			p.log.Errorw("Failed to process outbox message",
				"id", msg.ID,
				"aggregate_id", msg.AggregateID,
				"error", err,
			)
			outboxFailed.Inc()

			// Mark as failed
			if markErr := p.repo.MarkFailed(ctx, msg.ID, err); markErr != nil {
				p.log.Errorw("Failed to mark message as failed", "id", msg.ID, "error", markErr)
			}
		} else {
			outboxProcessed.Inc()
			p.log.Infow("Outbox message processed",
				"id", msg.ID,
				"aggregate_id", msg.AggregateID,
			)

			// Mark as processed
			if markErr := p.repo.MarkProcessed(ctx, msg.ID); markErr != nil {
				p.log.Errorw("Failed to mark message as processed", "id", msg.ID, "error", markErr)
			}
		}
	}
}

func (p *Processor) processMessage(ctx context.Context, msg *domain.OutboxMessage) error {
	var topic string

	switch msg.EventType {
	case domain.EventTypeOrderCreated:
		topic = events.TopicOrderCreated
	default:
		return fmt.Errorf("unknown event type: %s", msg.EventType)
	}

	kafkaMsg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(msg.AggregateID),
		Value: sarama.ByteEncoder(msg.Payload),
	}

	partition, offset, err := p.producer.SendMessage(kafkaMsg)
	if err != nil {
		return fmt.Errorf("failed to send message to Kafka: %w", err)
	}

	p.log.Infow("Message sent to Kafka",
		"topic", topic,
		"partition", partition,
		"offset", offset,
		"aggregate_id", msg.AggregateID,
	)

	return nil
}
