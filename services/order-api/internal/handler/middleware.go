package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/yourusername/go-event-order-processor/pkg/logger"
)

// CorrelationIDMiddleware adds correlation ID to request context
func CorrelationIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		correlationID := r.Header.Get("X-Correlation-ID")
		if correlationID == "" {
			correlationID = uuid.New().String()
		}

		// Add correlation ID to response headers
		w.Header().Set("X-Correlation-ID", correlationID)

		// Add correlation ID to context
		ctx := logger.WithCorrelationID(r.Context(), correlationID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
