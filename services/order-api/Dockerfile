FROM golang:1.23-alpine AS builder

WORKDIR /build

# Copy pkg modules first
COPY pkg/logger/go.mod pkg/logger/go.sum* /build/pkg/logger/
COPY pkg/events/go.mod pkg/events/go.sum* /build/pkg/events/
COPY pkg/idempotency/go.mod pkg/idempotency/go.sum* /build/pkg/idempotency/

# Copy service go.mod
COPY services/order-api/go.mod services/order-api/go.sum* /build/services/order-api/

WORKDIR /build/services/order-api
RUN go mod download || true

# Copy all source code
COPY pkg/ /build/pkg/
COPY services/order-api/ /build/services/order-api/

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/order-api ./cmd

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/order-api /app/order-api
COPY services/order-api/migrations /app/migrations

EXPOSE 8080

CMD ["/app/order-api"]
