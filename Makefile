.PHONY: help up down logs migrate-up migrate-down test build clean status

.DEFAULT_GOAL := help

help: ## Показать справку
	@echo "Доступные команды:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

up: ## Запустить все сервисы
	@echo "🚀 Запуск сервисов..."
	docker-compose -f deployments/docker-compose.yml up -d
	@echo "✅ Сервисы запущены!"
	@echo ""
	@echo "📍 Endpoints:"
	@echo "   order-api:      http://localhost:8080"
	@echo "   payment-worker: http://localhost:8081/metrics"
	@echo "   Prometheus:     http://localhost:9090"
	@echo "   Grafana:        http://localhost:3000 (admin/admin)"

down: ## Остановить все сервисы
	@echo "🛑 Остановка сервисов..."
	docker-compose -f deployments/docker-compose.yml down
	@echo "✅ Сервисы остановлены"

down-v: ## Остановить сервисы и удалить volumes
	@echo "🛑 Остановка сервисов и удаление данных..."
	docker-compose -f deployments/docker-compose.yml down -v
	@echo "✅ Сервисы остановлены, данные удалены"

logs: ## Показать логи всех сервисов
	docker-compose -f deployments/docker-compose.yml logs -f

logs-api: ## Показать логи order-api
	docker-compose -f deployments/docker-compose.yml logs -f order-api

logs-worker: ## Показать логи payment-worker
	docker-compose -f deployments/docker-compose.yml logs -f payment-worker

status: ## Проверить статус сервисов
	@echo "📊 Статус сервисов:"
	@docker-compose -f deployments/docker-compose.yml ps

migrate-up: ## Применить миграции БД
	@echo "🔧 Применение миграций..."
	docker run --rm --network=deployments_orders-network \
		-v "$(PWD)/services/order-api/migrations:/migrations" \
		migrate/migrate:latest \
		-path=/migrations \
		-database "postgres://postgres:postgres@postgres:5432/orders?sslmode=disable" \
		up
	@echo "✅ Миграции применены"

migrate-down: ## Откатить миграции БД
	@echo "🔧 Откат миграций..."
	docker run --rm --network=deployments_orders-network \
		-v "$(PWD)/services/order-api/migrations:/migrations" \
		migrate/migrate:latest \
		-path=/migrations \
		-database "postgres://postgres:postgres@postgres:5432/orders?sslmode=disable" \
		down
	@echo "✅ Миграции откачены"

test-api: ## Тестовый запрос создания заказа
	@echo "🧪 Создание тестового заказа..."
	@curl -X POST http://localhost:8080/api/v1/orders \
		-H "Content-Type: application/json" \
		-d '{"customer_id":"test-customer-123","amount":99.99}' \
		| jq .
	@echo ""
	@echo "✅ Проверьте логи payment-worker для обработки платежа:"
	@echo "   make logs-worker"

health: ## Проверить health check
	@echo "🏥 Проверка здоровья сервисов:"
	@echo -n "order-api: "
	@curl -s http://localhost:8080/health | jq -r '.status' 2>/dev/null || echo "❌ недоступен"

build: ## Собрать Docker образы
	@echo "🔨 Сборка образов..."
	docker-compose -f deployments/docker-compose.yml build
	@echo "✅ Образы собраны"

clean: ## Очистить Docker образы и volumes
	@echo "🧹 Очистка..."
	docker-compose -f deployments/docker-compose.yml down -v --rmi all
	@echo "✅ Очистка завершена"

restart: down up ## Перезапустить сервисы

rebuild: down build up ## Пересобрать и запустить

init: ## Первоначальная инициализация проекта
	@echo "🎬 Инициализация проекта..."
	@echo "1. Скачивание зависимостей..."
	cd pkg/logger && go mod tidy || true
	cd pkg/events && go mod tidy || true
	cd pkg/idempotency && go mod tidy || true
	cd services/order-api && go mod tidy || true
	cd services/payment-worker && go mod tidy || true
	@echo "2. Запуск инфраструктуры..."
	$(MAKE) up
	@echo "3. Ожидание готовности сервисов (30 сек)..."
	sleep 30
	@echo ""
	@echo "✅ Проект готов к работе!"
	@echo ""
	@echo "📖 Следующие шаги:"
	@echo "   1. Проверьте статус:     make status"
	@echo "   2. Создайте заказ:       make test-api"
	@echo "   3. Просмотрите логи:     make logs"
