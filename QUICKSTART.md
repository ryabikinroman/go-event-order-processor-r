# 🚀 Быстрый старт

## За 3 минуты до работающей системы

### 1. Подготовка

**Windows:**
```powershell
# Убедитесь, что Docker Desktop запущен
docker --version
docker-compose --version
```

**Linux/Mac:**
```bash
# Убедитесь, что установлены:
docker --version
docker-compose --version
make --version
```

### 2. Запуск

**Windows (PowerShell):**
```powershell
# Клонировать проект
git clone <repo-url>
cd go-event-order-processor

# Запустить всю инфраструктуру (одна команда!)
.\scripts\start.ps1

# Подождать 20-30 секунд для инициализации
```

**Linux/Mac:**
```bash
# Клонировать проект
git clone <repo-url>
cd go-event-order-processor

# Запустить всю инфраструктуру
make up

# Подождать 20-30 секунд для инициализации
```

### 3. Проверка

**Windows:**
```powershell
# Проверить статус сервисов
.\scripts\status.ps1
```

**Linux/Mac:**
```bash
make status
```

**Должны быть запущены:**
- ✅ order-api (8080)
- ✅ payment-worker (8081)
- ✅ postgres (5432)
- ✅ kafka (9092)
- ✅ redis (6379)
- ✅ prometheus (9090)
- ✅ grafana (3000)

### 4. Тестирование

#### Создать заказ

**Windows:**
```powershell
# Автоматический тест
.\scripts\test.ps1

# Или вручную:
curl -X POST http://localhost:8080/api/v1/orders -H "Content-Type: application/json" -d '{\"customer_id\":\"test-123\",\"amount\":99.99}'
```

**Linux/Mac:**
```bash
make test-api

# Или вручную:
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{"customer_id":"test-123","amount":99.99}'
```

#### Просмотреть логи обработки

**Windows:**
```powershell
# Логи payment-worker (увидите обработку платежа)
.\scripts\logs.ps1 payment-worker
```

**Linux/Mac:**
```bash
make logs-worker
```

**Должны увидеть:**
```json
{"level":"info","timestamp":"...","message":"Processing payment","order_id":"..."}
{"level":"info","timestamp":"...","message":"Payment processed successfully","order_id":"..."}
```

#### Проверить метрики

```bash
# Метрики order-api
curl http://localhost:8080/metrics | grep orders_created_total

# Метрики payment-worker
curl http://localhost:8081/metrics | grep payments_processed_total
```

## 📍 Доступные сервисы

| Сервис | URL | Описание |
|--------|-----|----------|
| **order-api** | http://localhost:8080 | REST API для заказов |
| **payment-worker** | http://localhost:8081/metrics | Метрики worker'а |
| **Prometheus** | http://localhost:9090 | Сбор метрик |
| **Grafana** | http://localhost:3000 | Дашборды (admin/admin) |
| PostgreSQL | localhost:5432 | БД (postgres/postgres) |
| Kafka | localhost:9092 | Message broker |
| Redis | localhost:6379 | Idempotency cache |

## 🛠️ Основные команды

### Windows (PowerShell)

```powershell
.\scripts\start.ps1      # Запустить сервисы
.\scripts\stop.ps1       # Остановить (данные сохраняются!)
.\scripts\status.ps1     # Статус контейнеров
.\scripts\logs.ps1       # Логи всех сервисов
.\scripts\test.ps1       # Тестовый запрос
.\scripts\rebuild.ps1    # Пересборка после изменений
.\scripts\clean.ps1      # Полная очистка

# Логи конкретного сервиса:
.\scripts\logs.ps1 order-api
.\scripts\logs.ps1 payment-worker
```

### Linux/Mac (Makefile)

```bash
make help          # Показать все команды
make up            # Запустить сервисы
make down          # Остановить сервисы
make logs          # Просмотреть логи
make logs-api      # Логи только order-api
make logs-worker   # Логи только payment-worker
make status        # Статус контейнеров
make test-api      # Тестовый запрос
make restart       # Перезапустить
make clean         # Очистить всё
```

## 🧪 Полный сценарий тестирования

**Windows:**
```powershell
# 1. Запустить систему
.\scripts\start.ps1

# 2. Подождать готовности (20-30 секунд)
Start-Sleep -Seconds 25

# 3. Создать несколько заказов
1..5 | ForEach-Object {
    $amount = Get-Random -Minimum 10 -Maximum 100
    curl -X POST http://localhost:8080/api/v1/orders `
      -H "Content-Type: application/json" `
      -d "{`"customer_id`":`"customer-$_`",`"amount`":$amount}"
}

# 4. Посмотреть обработку в логах
.\scripts\logs.ps1 payment-worker

# 5. Проверить метрики
curl -s http://localhost:8081/metrics | Select-String "payments_processed_total"

# 6. Посмотреть в Grafana
Start-Process http://localhost:3000  # admin/admin
```

**Linux/Mac:**
```bash
# 1. Запустить систему
make up

# 2. Подождать готовности
sleep 25

# 3. Создать несколько заказов
for i in {1..5}; do
  curl -X POST http://localhost:8080/api/v1/orders \
    -H "Content-Type: application/json" \
    -d "{\"customer_id\":\"customer-$i\",\"amount\":$(( RANDOM % 100 + 10 ))}"
  echo ""
done

# 4. Посмотреть обработку в логах
make logs-worker

# 5. Проверить метрики
curl -s http://localhost:8081/metrics | grep payments_processed_total

# 6. Посмотреть в Grafana
open http://localhost:3000  # admin/admin
```

## ⚠️ Troubleshooting

### Проблема: Kafka не запускается

**Решение:**
```powershell
# Полная очистка и перезапуск
docker-compose -f deployments/docker-compose.yml down -v
docker-compose -f deployments/docker-compose.yml up -d
```

### Проблема: Grafana не принимает пароль

**Решение:**
1. Подождите 10-20 секунд после запуска
2. Откройте http://localhost:3000 в **режиме инкогнито**
3. Логин: `admin`, Пароль: `admin`
4. Или перезапустите: `docker-compose -f deployments/docker-compose.yml restart grafana`

### Проблема: Порты заняты

**Windows:**
```powershell
# Проверить занятые порты
netstat -ano | findstr :8080
netstat -ano | findstr :5432

# Остановить всё
.\scripts\stop.ps1
```

**Linux/Mac:**
```bash
# Проверить занятые порты
lsof -i :8080
lsof -i :5432

# Остановить всё и очистить
make down-v
```

### Проблема: Сервисы не стартуют

```bash
# Посмотреть логи конкретного сервиса
docker-compose -f deployments/docker-compose.yml logs <service-name>

# Примеры:
docker-compose -f deployments/docker-compose.yml logs order-api
docker-compose -f deployments/docker-compose.yml logs kafka
docker-compose -f deployments/docker-compose.yml logs postgres
```

**Windows:**
```powershell
# Пересобрать образы
.\scripts\rebuild.ps1
```

**Linux/Mac:**
```bash
# Пересобрать образы
make rebuild
```

### Проблема: База данных не готова

```bash
# Проверить состояние PostgreSQL
docker exec orders-postgres pg_isready -U postgres

# Применить миграции вручную
make migrate-up
```

### Проблема: Kafka не принимает сообщения

```bash
# Проверить топики Kafka
docker exec orders-kafka kafka-topics \
  --bootstrap-server localhost:9092 \
  --list

# Должны быть:
# - order.created
# - order.dlq
```

## 📚 Следующие шаги

После успешного запуска:

### 1. Изучите код

**Основные файлы:**
- [services/order-api/internal/handler/http.go](services/order-api/internal/handler/http.go) - HTTP handlers
- [services/payment-worker/internal/consumer/kafka.go](services/payment-worker/internal/consumer/kafka.go) - Kafka consumer с retry logic
- [pkg/idempotency/idempotency.go](pkg/idempotency/idempotency.go) - Идемпотентность через Redis
- [pkg/logger/logger.go](pkg/logger/logger.go) - Структурированное логирование

### 2. Попробуйте сценарии отказа

**Тест отказоустойчивости:**
```bash
# 1. Создать заказ
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{"customer_id":"test-fail","amount":99.99}'

# 2. Остановить payment-worker
docker-compose -f deployments/docker-compose.yml stop payment-worker

# 3. Создать еще заказы (они будут ждать в Kafka)
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{"customer_id":"test-queue","amount":50.00}'

# 4. Запустить payment-worker снова
docker-compose -f deployments/docker-compose.yml start payment-worker

# 5. Наблюдать как заказы обработаются
docker-compose -f deployments/docker-compose.yml logs -f payment-worker
```

**Тест идемпотентности:**
- Обратите внимание на `order_id` в ответе
- Попробуйте отправить событие дважды (имитация retry)
- Payment-worker обработает его только 1 раз благодаря Redis

### 3. Изучите метрики

**Prometheus:**
- Откройте http://localhost:9090
- Попробуйте запросы:
  - `orders_created_total` - общее количество заказов
  - `payments_processed_total{status="success"}` - успешные платежи
  - `rate(http_request_duration_seconds_sum[5m])` - RPS

**Grafana:**
- Откройте http://localhost:3000 (admin/admin)
- Подключите Prometheus data source
- Создайте дашборд с метриками

### 4. Изучите архитектуру

Детальная документация:
- 📖 [README.md](README.md) - Полное описание проекта
- 🏗️ [ARCHITECTURE.md](ARCHITECTURE.md) - Архитектурные решения и паттерны

## 🔗 Полезные ссылки

- **Health Checks**:
  - order-api: http://localhost:8080/health

- **Метрики**:
  - order-api: http://localhost:8080/metrics
  - payment-worker: http://localhost:8081/metrics
  - Prometheus Targets: http://localhost:9090/targets

- **Мониторинг**:
  - Grafana Explore: http://localhost:3000/explore
  - Prometheus Graph: http://localhost:9090/graph

---

**Возникли вопросы?** Открывайте Issue в репозитории!

**Хотите улучшить?** Pull Requests приветствуются!
