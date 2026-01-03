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

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/yourusername/go-event-order-processor/pkg/logger"
	"github.com/yourusername/go-event-order-processor/services/order-api/internal/handler"
	"github.com/yourusername/go-event-order-processor/services/order-api/internal/producer"
	"github.com/yourusername/go-event-order-processor/services/order-api/internal/repository"
)

func main() {
	log, err := logger.New("order-api", getEnv("LOG_LEVEL", "info"))
	if err != nil {
		panic(err)
	}

	log.Info("Starting order-api service")

	// Connect to database
	dbConn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", "postgres"),
		getEnv("DB_NAME", "orders"),
	)

	db, err := repository.NewDB(dbConn)
	if err != nil {
		log.Fatalw("Failed to connect to database", "error", err)
	}
	defer db.Close()

	log.Info("Connected to PostgreSQL")

	// Initialize Kafka producer
	brokers := strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ",")
	kafkaProducer, err := producer.NewKafkaProducer(brokers, log)
	if err != nil {
		log.Fatalw("Failed to create Kafka producer", "error", err)
	}
	defer kafkaProducer.Close()

	log.Info("Kafka producer initialized")

	// Initialize handler
	repo := repository.NewPostgresRepository(db)
	h := handler.NewHandler(repo, kafkaProducer, log)
	h.SetHealthChecker(db)

	// Setup router
	r := mux.NewRouter()
	r.HandleFunc("/health", h.Health).Methods("GET")
	r.Handle("/metrics", promhttp.Handler()).Methods("GET")

	// Apply middleware to API routes
	api := r.PathPrefix("/api/v1").Subrouter()
	api.Use(handler.CorrelationIDMiddleware)
	api.HandleFunc("/orders", h.CreateOrder).Methods("POST")
	api.HandleFunc("/orders/{id}", h.GetOrder).Methods("GET")

	port := getEnv("PORT", "8080")
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Infow("Server started", "port", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalw("Server failed", "error", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Errorw("Server forced to shutdown", "error", err)
	}

	log.Info("Server exited")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
