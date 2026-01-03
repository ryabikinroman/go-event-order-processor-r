# Архитектура проекта

## Обзор

Go Event Order Processor - это демонстрационный проект, показывающий production-ready подход к построению event-driven систем на Go.

## Компоненты

### 1. order-api (HTTP REST API)

**Назначение**: Прием HTTP запросов, создание заказов, публикация событий.

**Технологии**:
- Gorilla Mux (роутинг)
- PostgreSQL (хранение заказов)
- Kafka Producer (публикация событий)
- Prometheus (метрики)

**Слои (Clean Architecture)**:
```
domain/      - Бизнес-сущности и правила (Order, OrderStatus)
handler/     - HTTP handlers, валидация запросов
repository/  - PostgreSQL репозиторий
producer/    - Kafka producer
```

**Основной flow**:
1. Принять POST `/api/v1/orders`
2. Валидировать данные (domain rules)
3. Сохранить в PostgreSQL
4. Опубликовать событие `order.created` в Kafka
5. Вернуть ответ клиенту

### 2. payment-worker (Kafka Consumer)

**Назначение**: Асинхронная обработка платежей с гарантией идемпотентности.

**Технологии**:
- Kafka Consumer Group
- Redis (idempotency checking)
- Prometheus (метрики)

**Основной flow**:
1. Потребить событие `order.created` из Kafka
2. Проверить идемпотентность через Redis (SetNX)
3. Если уже обработано - пропустить
4. Обработать платеж (симуляция)
5. При ошибке - retry с exponential backoff (3 попытки)
6. После исчерпания попыток - отправить в DLQ

**Retry Logic**:
- Попытка 1: сразу
- Попытка 2: через 2 секунды
- Попытка 3: через 4 секунды
- После этого → Dead Letter Queue

## Event Flow

```
┌─────────┐    POST /orders    ┌───────────┐
│ Client  │ ───────────────────▶│ order-api │
└─────────┘                     └─────┬─────┘
                                      │
                                      │ 1. Save to PostgreSQL
                                      │
                                      ▼
                            ┌──────────────────┐
                            │   PostgreSQL     │
                            └──────────────────┘
                                      │
                                      │ 2. Publish event
                                      ▼
                                ┌───────────┐
                                │   Kafka   │
                                │  order.   │
                                │  created  │
                                └─────┬─────┘
                                      │
                                      │ 3. Consume
                                      ▼
                          ┌────────────────────┐
                          │  payment-worker    │
                          └─────┬──────────────┘
                                │
                    ┌───────────┼───────────┐
                    │           │           │
              4a. Check   4b. Process  4c. On failure
              idempotency   payment     (after retries)
                    │           │           │
                    ▼           ▼           ▼
              ┌─────────┐  Success!   ┌─────────┐
              │  Redis  │              │   DLQ   │
              └─────────┘              └─────────┘
```

## Гарантии надежности

### Идемпотентность
- **Проблема**: Kafka может доставить сообщение >1 раза
- **Решение**: Redis SetNX с TTL 24 часа
- **Ключ**: `processed:{event_id}`
- **Результат**: Даже при повторной доставке платеж обработается только 1 раз

### Retry с Exponential Backoff
- **Зачем**: Временные сбои (сеть, внешний API)
- **Реализация**: 3 попытки с задержками 0s, 2s, 4s
- **Dead Letter Queue**: Для дальнейшего расследования

### Graceful Shutdown
- Сервисы корректно завершают обработку при SIGTERM/SIGINT
- Kafka consumer закрывает соединения
- HTTP сервер дожидается завершения запросов

## Observability

### Метрики (Prometheus)

**order-api**:
- `orders_created_total` - количество созданных заказов
- `http_request_duration_seconds` - latency HTTP запросов

**payment-worker**:
- `payments_processed_total{status}` - обработанные платежи (success/failed)
- `messages_consumed_total` - потребленные сообщения
- `messages_retried_total` - повторные попытки
- `messages_sent_to_dlq_total` - сообщения в DLQ
- `payment_processing_duration_seconds` - время обработки

### Логирование

Структурированные JSON логи (zap):
```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "level": "info",
  "service": "payment-worker",
  "message": "Payment processed successfully",
  "event_id": "abc-123",
  "order_id": "ord-456",
  "correlation_id": "xyz-789"
}
```

**Correlation ID**: Прослеживается через все сервисы для distributed tracing.

## Production Best Practices

✅ **Clean Architecture** - явное разделение слоев
✅ **Domain-Driven Design** - бизнес-логика в domain layer
✅ **Repository Pattern** - абстракция БД
✅ **Dependency Injection** - через конструкторы
✅ **Idempotency** - Redis для дедупликации
✅ **Retry Logic** - с exponential backoff
✅ **Dead Letter Queue** - для неудачных сообщений
✅ **Structured Logging** - JSON с correlation IDs
✅ **Metrics** - Prometheus + Grafana
✅ **Health Checks** - для orchestrators
✅ **Graceful Shutdown** - корректное завершение
✅ **Database Migrations** - версионирование схемы
✅ **Configuration via Env** - 12-factor app

## Масштабирование

### Горизонтальное масштабирование
- **order-api**: Stateless, можно запустить N инстансов за load balancer
- **payment-worker**: Kafka Consumer Group автоматически распределяет партиции между инстансами

### Вертикальное масштабирование
- PostgreSQL: connection pooling (max 25 connections)
- Kafka: 3 партиции для параллельной обработки

## Ограничения и Trade-offs

### Eventual Consistency
- Заказ сначала сохраняется, потом обрабатывается платеж
- Возможно состояние "заказ создан, платеж еще не обработан"

### Нет транзакции между DB и Kafka
- Outbox pattern не реализован для простоты
- Production решение: Transactional Outbox или CDC (Debezium)

### Симуляция платежей
- Реальный payment gateway не интегрирован
- 10% платежей случайно "фейлятся" для демонстрации retry logic

## Тестирование

### Unit тесты
- Domain logic
- Business rules validation

### Integration тесты
- Полный flow с реальными PostgreSQL/Kafka/Redis через testcontainers

### Load тесты (опционально)
```bash
# Apache Bench
ab -n 1000 -c 10 -p order.json -T application/json \
  http://localhost:8080/api/v1/orders
```

## Дальнейшие улучшения

Возможные доработки для production:
- [ ] Distributed tracing (OpenTelemetry/Jaeger)
- [ ] Circuit Breaker (для внешних API)
- [ ] Rate Limiting
- [ ] API Gateway (nginx/Kong)
- [ ] Kubernetes deployment
- [ ] Helm charts
- [ ] E2E тесты
- [ ] Outbox pattern для гарантий доставки
- [ ] Schema Registry для Kafka events

## Development Workflow

Для упрощения разработки и тестирования проект включает helper scripts в папке `scripts/`:

### Windows (PowerShell)
```powershell
.\scripts\start.ps1      # Запуск всех сервисов
.\scripts\stop.ps1       # Остановка (данные сохраняются)
.\scripts\status.ps1     # Статус контейнеров
.\scripts\logs.ps1       # Просмотр логов
.\scripts\test.ps1       # Тестирование API
.\scripts\rebuild.ps1    # Пересборка после изменений кода
.\scripts\clean.ps1      # Полная очистка (с подтверждением)
```

### Linux/Mac
```bash
make up          # Запуск
make down        # Остановка
make status      # Статус
make logs        # Логи
make rebuild     # Пересборка
```

Подробнее см. [QUICKSTART.md](QUICKSTART.md)
