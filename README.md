# Go Event Order Processor

> **Production-ready event-driven система обработки заказов**

[![Go Version](https://img.shields.io/badge/Go-1.23-blue.svg)](https://golang.org)

## 🎯 О проекте

Демонстрационный проект, показывающий **production-ready** подход к построению микросервисов на Go с event-driven архитектурой.

**Философия**: Качество кода важнее количества функций.

### Что реализовано

- ✅ **2 микросервиса** с четким разделением ответственности
- ✅ **Event-Driven Architecture** через Kafka
- ✅ **Clean Architecture** с явными слоями
- ✅ **Идемпотентность** через Redis
- ✅ **Retry logic** с exponential backoff + Dead Letter Queue
- ✅ **Observability**: Prometheus + Grafana + структурированные логи
- ✅ **Production patterns**: graceful shutdown, health checks, миграции

---

## 🏗️ Архитектура

```
┌──────────┐    HTTP     ┌─────────────┐
│  Client  │────────────▶│  order-api  │
└──────────┘             └──────┬──────┘
                                │
                        1. Save │ PostgreSQL
                        2. Publish
                                ▼
                          ┌──────────┐
                          │  Kafka   │
                          └────┬─────┘
                               │ order.created
                               ▼
                    ┌──────────────────────┐
                    │  payment-worker      │
                    └──────────────────────┘
                         │          │
                    3. Check   4. Process
                    Redis      Payment
                         │          │
                      ┌──┴──┐   Success/
                      │ ✓/✗ │   3 retries
                      └─────┘       │
                                    ▼ (on failure)
                               ┌─────────┐
                               │   DLQ   │
                               └─────────┘
```

### Сервисы

#### 1. **order-api** (порт 8080)
- Прием и сохранение заказов
- Публикация событий в Kafka
- **Endpoints**: `POST /api/v1/orders`, `GET /api/v1/orders/:id`, `/health`, `/metrics`

#### 2. **payment-worker** (порт 8081)
- Асинхронная обработка платежей
- Идемпотентная обработка через Redis
- Retry logic (3 попытки с exponential backoff)
- Dead Letter Queue для failed сообщений

---

## 🚀 Быстрый старт

### Требования
- Docker Desktop для Windows
- Go 1.23+ (для локальной разработки)
- Make (опционально, для Linux/Mac)

### Запуск

**Для Windows (PowerShell):**

```powershell
# 1. Клонировать репозиторий
git clone <repo-url>
cd go-event-order-processor

# 2. Запустить проект (одна команда!)
.\scripts\start.ps1
```

**Для Linux/Mac:**

```bash
git clone <repo-url>
cd go-event-order-processor
make up
```

Готово! 🎉

**Доступные сервисы:**
- order-api: http://localhost:8080
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000 (admin/admin)

### Тестирование

**Windows (PowerShell):**
```powershell
# Создать тестовый заказ (всё в одном скрипте)
.\scripts\test.ps1

# Или вручную:
curl -X POST http://localhost:8080/api/v1/orders -H "Content-Type: application/json" -d '{\"customer_id\":\"test-123\",\"amount\":99.99}'

# Посмотреть логи worker
.\scripts\logs.ps1 payment-worker

# Проверить метрики
curl http://localhost:8081/metrics | Select-String "payments_processed_total"
```

**Linux/Mac:**
```bash
# Создать заказ
make test-api

# Посмотреть обработку
make logs-worker

# Проверить метрики
curl http://localhost:8081/metrics | grep payments_processed_total
```

---

## 📡 Использование API

### Создать заказ
```bash
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{"customer_id":"cust-123","amount":99.99}'
```

**Ответ:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "customer_id": "cust-123",
  "amount": 99.99,
  "currency": "USD",
  "status": "pending",
  "created_at": "2024-01-15T10:30:00Z"
}
```

### Получить заказ
```bash
curl http://localhost:8080/api/v1/orders/{order_id}
```

### Health Check
```bash
curl http://localhost:8080/health
# {"status":"healthy"}
```

---

## 📁 Структура проекта

```
go-event-order-processor/
├── services/
│   ├── order-api/              # HTTP REST API
│   │   ├── cmd/                # Entry point
│   │   ├── internal/
│   │   │   ├── domain/         # Бизнес-логика
│   │   │   ├── handler/        # HTTP handlers
│   │   │   ├── repository/     # PostgreSQL
│   │   │   └── producer/       # Kafka producer
│   │   ├── migrations/         # DB migrations
│   │   └── Dockerfile
│   │
│   └── payment-worker/         # Kafka consumer
│       ├── cmd/
│       ├── internal/
│       │   ├── consumer/       # Kafka consumer
│       │   └── processor/      # Payment logic
│       └── Dockerfile
│
├── pkg/                        # Shared код
│   ├── logger/                 # Структурированные логи
│   ├── events/                 # Event schemas
│   └── idempotency/            # Redis идемпотентность
│
├── deployments/
│   ├── docker-compose.yml
│   └── prometheus.yml
│
├── Makefile
├── README.md
│
├── scripts/                       # PowerShell скрипты (для Windows)
│   ├── start.ps1                  # Запуск проекта
│   ├── stop.ps1                   # Остановка
│   ├── status.ps1                 # Статус контейнеров
│   ├── logs.ps1                   # Просмотр логов
│   ├── test.ps1                   # Тестирование API
│   ├── rebuild.ps1                # Пересборка после изменений
│   └── clean.ps1                  # Полная очистка
│
├── Makefile
└── README.md
```

---

## 🛠️ Команды управления

### Для Windows (PowerShell):

**Простые команды:**

```powershell
.\scripts\start.ps1      # Запустить всё
.\scripts\stop.ps1       # Остановить (данные сохраняются)
.\scripts\status.ps1     # Проверить статус
.\scripts\logs.ps1       # Логи всех сервисов
.\scripts\test.ps1       # Создать тестовый заказ

# Логи конкретного сервиса:
.\scripts\logs.ps1 order-api
.\scripts\logs.ps1 payment-worker

# После изменения кода:
.\scripts\rebuild.ps1    # Пересборка и перезапуск

# Полная очистка (удаляет все данные):
.\scripts\clean.ps1
```

**Или используйте docker-compose напрямую:**

```powershell
# Запустить
docker-compose -f deployments/docker-compose.yml up -d

# Остановить (данные сохраняются)
docker-compose -f deployments/docker-compose.yml stop

# Остановить и удалить контейнеры (данные в volumes сохраняются)
docker-compose -f deployments/docker-compose.yml down

# Полная очистка (удаляет volumes)
docker-compose -f deployments/docker-compose.yml down -v
```

### Для Linux/Mac (Makefile):

```bash
make help          # Показать все команды
make up            # Запустить всё
make down          # Остановить
make logs          # Логи всех сервисов
make logs-api      # Логи только order-api
make logs-worker   # Логи только payment-worker
make status        # Статус контейнеров
make test-api      # Тестовый запрос
make health        # Health check
make restart       # Перезапустить
make clean         # Удалить всё
```

---

## 🎯 Production-Ready паттерны

### Архитектура
- ✅ **Clean Architecture** (domain, usecase, infrastructure, delivery)
- ✅ **Repository Pattern** - абстракция БД
- ✅ **Dependency Injection** - через конструкторы
- ✅ **Domain-Driven Design** - бизнес-логика в domain layer

### Надежность
- ✅ **Идемпотентность** через Redis SetNX
- ✅ **Retry logic** с exponential backoff (3 попытки: 0s, 2s, 4s)
- ✅ **Dead Letter Queue** для failed сообщений
- ✅ **Graceful shutdown** всех сервисов
- ✅ **Database transactions**

### Observability
- ✅ **Structured JSON logs** с correlation IDs (zap)
- ✅ **Prometheus metrics** (business + technical)
- ✅ **Health checks** для orchestrators
- ✅ **Grafana** для визуализации

### DevOps
- ✅ **Docker** multi-stage builds
- ✅ **Docker Compose** с health checks
- ✅ **Database migrations** автоматические
- ✅ **Environment-based config** (12-factor app)

---

## 📊 Метрики

### order-api
- `orders_created_total` - количество созданных заказов
- `http_request_duration_seconds` - latency HTTP запросов

### payment-worker
- `payments_processed_total{status}` - обработанные платежи (success/failed)
- `messages_consumed_total` - потребленные сообщения
- `messages_retried_total` - повторные попытки
- `messages_sent_to_dlq_total` - сообщения в DLQ
- `payment_processing_duration_seconds` - время обработки

---

## 🧪 Тестовые сценарии

### 1. Happy Path
```bash
make test-api
make logs-worker  # Увидите успешную обработку
```

### 2. Идемпотентность
```bash
# payment-worker использует Redis для предотвращения дубликатов
# Даже при повторной доставке Kafka, платеж обработается только 1 раз
```

### 3. Retry & DLQ
```bash
# 10% платежей рандомно фейлятся (для демонстрации)
# Наблюдайте retry logic в логах:
make logs-worker

# После 3 неудачных попыток → DLQ
curl http://localhost:8081/metrics | grep messages_sent_to_dlq_total
```

---

## 📚 Документация

- **[QUICKSTART.md](QUICKSTART.md)** - Быстрый старт за 3 минуты
- **[ARCHITECTURE.md](ARCHITECTURE.md)** - Детальная архитектура и design decisions

---

## 🔑 Технологический стек

| Технология | Назначение |
|------------|------------|
| Go 1.23 | Основной язык |
| PostgreSQL | Хранение заказов |
| Kafka | Event streaming |
| Redis | Idempotency cache |
| Prometheus | Сбор метрик |
| Grafana | Визуализация метрик |
| Docker | Контейнеризация |
| Zap | Структурированное логирование |
| Gorilla Mux | HTTP роутинг |

---

## 🚧 Возможные улучшения

Идеи для расширения проекта:
- [ ] Unit tests (testify/mock)
- [ ] OpenAPI/Swagger спецификация
- [ ] Distributed tracing (OpenTelemetry)
- [ ] Circuit Breaker pattern
- [ ] Rate limiting
- [ ] Outbox pattern для transactional guarantees
- [ ] Kubernetes manifests

---

## 👨‍💻 Автор

ryabikinroman33@gmail.com

**Принципы разработки**:
- Качество > количество
- Production-ready > feature-rich
- Простота > сложность
---

**⭐ Если проект был полезен, поставьте звезду!**

**📧 Вопросы?** Открывайте Issue в репозитории!
