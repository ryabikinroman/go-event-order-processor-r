package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/yourusername/go-event-order-processor/pkg/logger"
	"github.com/yourusername/go-event-order-processor/services/order-api/internal/domain"
	"github.com/yourusername/go-event-order-processor/services/order-api/internal/producer"
)

var (
	ordersCreated = promauto.NewCounter(prometheus.CounterOpts{
		Name: "orders_created_total",
		Help: "Total number of orders created",
	})

	httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "http_request_duration_seconds",
		Help: "HTTP request duration in seconds",
	}, []string{"path", "method"})
)

type Handler struct {
	repo     domain.OrderRepository
	producer *producer.KafkaProducer
	log      *logger.Logger
}

func NewHandler(repo domain.OrderRepository, producer *producer.KafkaProducer, log *logger.Logger) *Handler {
	return &Handler{
		repo:     repo,
		producer: producer,
		log:      log,
	}
}

type CreateOrderRequest struct {
	CustomerID string  `json:"customer_id"`
	Amount     float64 `json:"amount"`
}

type OrderResponse struct {
	ID         string  `json:"id"`
	CustomerID string  `json:"customer_id"`
	Amount     float64 `json:"amount"`
	Currency   string  `json:"currency"`
	Status     string  `json:"status"`
	CreatedAt  string  `json:"created_at"`
}

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	timer := prometheus.NewTimer(httpDuration.WithLabelValues("/api/v1/orders", "POST"))
	defer timer.ObserveDuration()

	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	order, err := domain.NewOrder(req.CustomerID, req.Amount)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	order.ID = uuid.New().String()

	if err := h.repo.Create(r.Context(), order); err != nil {
		h.log.Errorw("Failed to create order", "error", err)
		h.respondError(w, http.StatusInternalServerError, "failed to create order")
		return
	}

	if err := h.producer.PublishOrderCreated(r.Context(), order); err != nil {
		h.log.Errorw("Failed to publish event", "error", err)
		// Не возвращаем ошибку пользователю, т.к. заказ уже создан
	}

	ordersCreated.Inc()

	h.respondJSON(w, http.StatusCreated, h.toResponse(order))
}

func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	timer := prometheus.NewTimer(httpDuration.WithLabelValues("/api/v1/orders/{id}", "GET"))
	defer timer.ObserveDuration()

	vars := mux.Vars(r)
	id := vars["id"]

	order, err := h.repo.GetByID(r.Context(), id)
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

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(`{"status":"healthy"}`)); err != nil {
		h.log.Errorw("Failed to write health response", "error", err)
	}
}

func (h *Handler) toResponse(order *domain.Order) OrderResponse {
	return OrderResponse{
		ID:         order.ID,
		CustomerID: order.CustomerID,
		Amount:     order.Amount,
		Currency:   order.Currency,
		Status:     string(order.Status),
		CreatedAt:  order.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func (h *Handler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.log.Errorw("Failed to encode JSON response", "error", err)
	}
}

func (h *Handler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{"error": message})
}
