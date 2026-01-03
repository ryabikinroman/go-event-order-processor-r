package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/IBM/sarama"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/yourusername/go-event-order-processor/pkg/events"
	"github.com/yourusername/go-event-order-processor/pkg/logger"
	"github.com/yourusername/go-event-order-processor/services/payment-worker/internal/processor"
)

var (
	messagesConsumed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "messages_consumed_total",
		Help: "Total number of messages consumed",
	})

	messagesRetried = promauto.NewCounter(prometheus.CounterOpts{
		Name: "messages_retried_total",
		Help: "Total number of messages retried",
	})

	messagesSentToDLQ = promauto.NewCounter(prometheus.CounterOpts{
		Name: "messages_sent_to_dlq_total",
		Help: "Total number of messages sent to DLQ",
	})
)

type KafkaConsumer struct {
	consumer     sarama.ConsumerGroup
	producer     sarama.SyncProducer
	processor    *processor.Processor
	log          *logger.Logger
	maxRetries   int
	retryBackoff time.Duration
}

func NewKafkaConsumer(
	brokers []string,
	groupID string,
	processor *processor.Processor,
	log *logger.Logger,
) (*KafkaConsumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}

	consumer, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer group: %w", err)
	}

	// Producer для DLQ
	producerConfig := sarama.NewConfig()
	producerConfig.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(brokers, producerConfig)
	if err != nil {
		consumer.Close()
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	return &KafkaConsumer{
		consumer:     consumer,
		producer:     producer,
		processor:    processor,
		log:          log,
		maxRetries:   3,
		retryBackoff: time.Second,
	}, nil
}

func (c *KafkaConsumer) Start(ctx context.Context) error {
	topics := []string{events.TopicOrderCreated}

	handler := &consumerGroupHandler{
		consumer: c,
	}

	for {
		if err := c.consumer.Consume(ctx, topics, handler); err != nil {
			return fmt.Errorf("error from consumer: %w", err)
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
}

func (c *KafkaConsumer) Close() error {
	if err := c.consumer.Close(); err != nil {
		return err
	}
	return c.producer.Close()
}

type consumerGroupHandler struct {
	consumer *KafkaConsumer
}

func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		messagesConsumed.Inc()

		if err := h.processMessage(session.Context(), message); err != nil {
			h.consumer.log.Errorw("Failed to process message", "error", err)
			continue
		}

		session.MarkMessage(message, "")
	}

	return nil
}

func (h *consumerGroupHandler) processMessage(ctx context.Context, msg *sarama.ConsumerMessage) error {
	var event events.OrderCreatedEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	h.consumer.log.Infow("Received message",
		"event_id", event.Metadata.EventID,
		"order_id", event.OrderID,
	)

	// Retry logic с exponential backoff
	var lastErr error
	for attempt := 0; attempt < h.consumer.maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(math.Pow(2, float64(attempt))) * h.consumer.retryBackoff
			h.consumer.log.Infow("Retrying message",
				"attempt", attempt+1,
				"backoff", backoff,
			)
			messagesRetried.Inc()
			time.Sleep(backoff)
		}

		err := h.consumer.processor.ProcessPayment(ctx, event)
		if err == nil {
			return nil
		}

		lastErr = err
		h.consumer.log.Warnw("Processing attempt failed",
			"attempt", attempt+1,
			"error", err,
		)
	}

	// Отправка в DLQ после исчерпания попыток
	h.consumer.log.Errorw("Max retries exceeded, sending to DLQ",
		"event_id", event.Metadata.EventID,
		"order_id", event.OrderID,
		"error", lastErr,
	)

	if err := h.sendToDLQ(msg); err != nil {
		h.consumer.log.Errorw("Failed to send to DLQ", "error", err)
	}

	messagesSentToDLQ.Inc()

	return lastErr
}

func (h *consumerGroupHandler) sendToDLQ(originalMsg *sarama.ConsumerMessage) error {
	dlqMsg := &sarama.ProducerMessage{
		Topic: events.TopicOrderDLQ,
		Key:   sarama.ByteEncoder(originalMsg.Key),
		Value: sarama.ByteEncoder(originalMsg.Value),
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("original_topic"),
				Value: []byte(originalMsg.Topic),
			},
			{
				Key:   []byte("failed_at"),
				Value: []byte(time.Now().Format(time.RFC3339)),
			},
		},
	}

	_, _, err := h.consumer.producer.SendMessage(dlqMsg)
	return err
}
