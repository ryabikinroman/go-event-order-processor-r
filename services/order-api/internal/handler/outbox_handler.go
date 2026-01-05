package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yourusername/go-event-order-processor/pkg/events"
	"github.com/yourusername/go-event-order-processor/pkg/logger"
	"github.com/yourusername/go-event-order-processor/services/order-api/internal/domain"
)

type OrderRepository interface {
	CreateWithOutbox(ctx context.Context, order *domain.Order, outboxMsg *domain.OutboxMessage) error
	GetByID(ctx context.Context, id string) (*domain.Order, error)
}

type OutboxHandler struct {
	repo OrderRepository
	log  *logger.Logger
	db   HealthChecker
}

func NewOutboxHandler(repo OrderRepository, log *logger.Logger) *OutboxHandler {
	return &OutboxHandler{
		repo: repo,
		log:  log,
	}
}

func (h *OutboxHandler) SetHealthChecker(db HealthChecker) {
	h.db = db
}

func (h *OutboxHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	timer := prometheus.NewTimer(httpDuration.WithLabelValues("/api/v1/orders", "POST"))
	defer timer.ObserveDuration()

	// Add 5-second timeout
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Limit request body size to prevent DoS attacks (1MB limit)
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := req.Validate(); err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	order, err := domain.NewOrder(req.CustomerID, req.Amount)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	order.ID = uuid.New().String()

	// Create outbox message
	event := events.OrderCreatedEvent{
		Metadata: events.EventMetadata{
			EventID:       uuid.New().String(),
			CorrelationID: getCorrelationID(ctx),
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
		h.log.Errorw("Failed to marshal event", "error", err)
		h.respondError(w, http.StatusInternalServerError, "failed to create order")
		return
	}

	outboxMsg := &domain.OutboxMessage{
		AggregateID:   order.ID,
		AggregateType: domain.AggregateTypeOrder,
		EventType:     domain.EventTypeOrderCreated,
		Payload:       payload,
		CreatedAt:     time.Now(),
		RetryCount:    0,
	}

	// Save order and outbox message in transaction
	if err := h.repo.CreateWithOutbox(ctx, order, outboxMsg); err != nil {
		h.log.Errorw("Failed to create order with outbox", "error", err)
		h.respondError(w, http.StatusInternalServerError, "failed to create order")
		return
	}

	ordersCreated.Inc()
	h.log.Infow("Order created successfully",
		"order_id", order.ID,
		"correlation_id", event.Metadata.CorrelationID,
	)

	h.respondJSON(w, http.StatusCreated, h.toResponse(order))
}

func (h *OutboxHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	timer := prometheus.NewTimer(httpDuration.WithLabelValues("/api/v1/orders/{id}", "GET"))
	defer timer.ObserveDuration()

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	vars := mux.Vars(r)
	id := vars["id"]

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid order id format")
		return
	}

	order, err := h.repo.GetByID(ctx, id)
	if err == domain.ErrOrderNotFound {
		h.respondError(w, http.StatusNotFound, "order not found")
		return
	}
	if err != nil {
		h.log.Errorw("Failed to get order", "error", err)
		h.respondError(w, http.StatusInternalServerError, "failed to get order")
		return
	}

	h.respondJSON(w, http.StatusOK, h.toResponse(order))
}

func (h *OutboxHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Check database connectivity
	if h.db != nil {
		if err := h.db.Ping(); err != nil {
			h.log.Errorw("Database health check failed", "error", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			if _, writeErr := w.Write([]byte(`{"status":"unhealthy","reason":"database connection failed"}`)); writeErr != nil {
				h.log.Errorw("Failed to write health response", "error", writeErr)
			}
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(`{"status":"healthy"}`)); err != nil {
		h.log.Errorw("Failed to write health response", "error", err)
	}
}

func (h *OutboxHandler) toResponse(order *domain.Order) OrderResponse {
	return OrderResponse{
		ID:         order.ID,
		CustomerID: order.CustomerID,
		Amount:     order.Amount,
		Currency:   order.Currency,
		Status:     string(order.Status),
		CreatedAt:  order.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func (h *OutboxHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.log.Errorw("Failed to encode JSON response", "error", err)
	}
}

func (h *OutboxHandler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{"error": message})
}

func getCorrelationID(ctx context.Context) string {
	type contextKey string
	const correlationIDKey contextKey = "correlation_id"

	if id, ok := ctx.Value(correlationIDKey).(string); ok {
		return id
	}
	return uuid.New().String()
}
