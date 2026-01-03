package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"github.com/yourusername/go-event-order-processor/pkg/idempotency"
	"github.com/yourusername/go-event-order-processor/pkg/logger"
	"github.com/yourusername/go-event-order-processor/services/payment-worker/internal/consumer"
	"github.com/yourusername/go-event-order-processor/services/payment-worker/internal/processor"
)

func main() {
	log, err := logger.New("payment-worker", getEnv("LOG_LEVEL", "info"))
	if err != nil {
		panic(err)
	}

	log.Info("Starting payment-worker service")

	// Initialize Redis
	redisAddr := fmt.Sprintf("%s:%s",
		getEnv("REDIS_HOST", "localhost"),
		getEnv("REDIS_PORT", "6379"),
	)

	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalw("Failed to connect to Redis", "error", err)
	}
	defer redisClient.Close()

	log.Info("Connected to Redis")

	// Initialize idempotency checker (TTL 24 hours)
	idempotencyChecker := idempotency.NewChecker(redisClient, 24*time.Hour)

	// Initialize processor
	proc := processor.NewProcessor(idempotencyChecker, log)

	// Initialize Kafka consumer
	brokers := strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ",")
	groupID := getEnv("KAFKA_GROUP_ID", "payment-workers")

	kafkaConsumer, err := consumer.NewKafkaConsumer(brokers, groupID, proc, log)
	if err != nil {
		log.Fatalw("Failed to create Kafka consumer", "error", err)
	}
	defer kafkaConsumer.Close()

	log.Info("Kafka consumer initialized")

	// Start metrics server
	metricsPort := getEnv("METRICS_PORT", "8081")
	metricsServer := &http.Server{
		Addr:         ":" + metricsPort,
		Handler:      promhttp.Handler(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Infow("Metrics server started", "port", metricsPort)
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Errorw("Metrics server failed", "error", err)
		}
	}()

	// Start consumer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := kafkaConsumer.Start(ctx); err != nil {
			log.Fatalw("Consumer failed", "error", err)
		}
	}()

	log.Info("Consumer started")

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down worker...")
	cancel()

	// Shutdown metrics server gracefully
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := metricsServer.Shutdown(shutdownCtx); err != nil {
		log.Errorw("Metrics server forced to shutdown", "error", err)
	}

	// Give some time for consumer to finish processing
	time.Sleep(5 * time.Second)

	log.Info("Worker exited")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
