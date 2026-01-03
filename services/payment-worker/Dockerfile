FROM golang:1.23-alpine AS builder

WORKDIR /build

# Copy pkg modules first
COPY pkg/logger/go.mod pkg/logger/go.sum* /build/pkg/logger/
COPY pkg/events/go.mod pkg/events/go.sum* /build/pkg/events/
COPY pkg/idempotency/go.mod pkg/idempotency/go.sum* /build/pkg/idempotency/

# Copy service go.mod
COPY services/payment-worker/go.mod services/payment-worker/go.sum* /build/services/payment-worker/

WORKDIR /build/services/payment-worker
RUN go mod download || true

# Copy all source code
COPY pkg/ /build/pkg/
COPY services/payment-worker/ /build/services/payment-worker/

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/payment-worker ./cmd

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/payment-worker /app/payment-worker

EXPOSE 8081

CMD ["/app/payment-worker"]
