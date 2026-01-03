package processor

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/yourusername/go-event-order-processor/pkg/events"
	"github.com/yourusername/go-event-order-processor/pkg/idempotency"
	"github.com/yourusername/go-event-order-processor/pkg/logger"
)

var (
	paymentsProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "payments_processed_total",
		Help: "Total number of payments processed",
	}, []string{"status"})

	paymentDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name: "payment_processing_duration_seconds",
		Help: "Payment processing duration in seconds",
	})
)

type Processor struct {
	idempotency *idempotency.Checker
	log         *logger.Logger
}

func NewProcessor(idempotency *idempotency.Checker, log *logger.Logger) *Processor {
	return &Processor{
		idempotency: idempotency,
		log:         log,
	}
}

func (p *Processor) ProcessPayment(ctx context.Context, event events.OrderCreatedEvent) error {
	timer := prometheus.NewTimer(paymentDuration)
	defer timer.ObserveDuration()

	// Проверка идемпотентности
	isNew, err := p.idempotency.CheckAndSet(ctx, event.Metadata.EventID)
	if err != nil {
		return fmt.Errorf("idempotency check failed: %w", err)
	}

	if !isNew {
		p.log.Infow("Event already processed, skipping",
			"event_id", event.Metadata.EventID,
			"order_id", event.OrderID,
		)
		return nil
	}

	p.log.Infow("Processing payment",
		"event_id", event.Metadata.EventID,
		"order_id", event.OrderID,
		"amount", event.Amount,
	)

	// Симуляция обработки платежа
	time.Sleep(time.Duration(rand.IntN(500)+100) * time.Millisecond)

	// Симуляция случайных ошибок (10% вероятность)
	if rand.IntN(10) == 0 {
		paymentsProcessed.WithLabelValues("failed").Inc()
		return fmt.Errorf("payment gateway error: random failure")
	}

	// Успешная обработка
	p.log.Infow("Payment processed successfully",
		"event_id", event.Metadata.EventID,
		"order_id", event.OrderID,
	)

	paymentsProcessed.WithLabelValues("success").Inc()
	return nil
}
