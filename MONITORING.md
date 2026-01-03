# Мониторинг и Метрики

## Обзор

Проект использует Prometheus для сбора метрик и Grafana для их визуализации.

## Доступ к сервисам

После запуска `.\scripts\start.ps1`:

- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000
  - Логин: `admin`
  - Пароль: `admin`

## Что смотреть после тестовой покупки

### 1. Prometheus (http://localhost:9090)

В поле поиска метрик вводите следующие запросы:

#### Основные метрики:

```promql
# Количество созданных заказов
orders_created_total

# Количество обработанных сообщений
messages_consumed_total

# Успешные платежи
payments_processed_total{status="success"}

# Неудачные платежи
payments_processed_total{status="failed"}

# Повторные попытки
messages_retried_total

# Сообщения в DLQ (после всех попыток)
messages_sent_to_dlq_total
```

#### Метрики производительности:

```promql
# Время обработки платежа (95-й перцентиль)
histogram_quantile(0.95, rate(payment_processing_duration_seconds_bucket[5m]))

# Скорость обработки заказов (в минуту)
rate(orders_created_total[1m]) * 60

# Процент успешных платежей
rate(payments_processed_total{status="success"}[5m]) / rate(payments_processed_total[5m]) * 100
```

### 2. Grafana (http://localhost:3000)

После входа в Grafana вы увидите автоматически настроенный dashboard **"Order Processor Dashboard"**.

#### Dashboard содержит:

1. **Total Orders Created** - общее количество созданных заказов
2. **Messages Consumed** - сколько сообщений обработал payment-worker
3. **Payment Processing Rate** - график скорости обработки (успешные/неудачные)
4. **Payment Processing Duration** - время обработки платежей (p50, p95)
5. **Successful Payments** - счетчик успешных платежей
6. **Messages Retried** - сколько сообщений были повторно обработаны
7. **Messages in DLQ** - сколько сообщений попало в Dead Letter Queue

#### Как открыть dashboard:

1. Откройте http://localhost:3000
2. Войдите (admin/admin)
3. Нажмите на иконку меню (☰) слева
4. Выберите "Dashboards"
5. Найдите "Order Processor Dashboard"

## Сценарий тестирования

### 1. Создайте несколько заказов:

```powershell
# Создать 5 тестовых заказов
1..5 | ForEach-Object { .\scripts\test.ps1; Start-Sleep -Seconds 1 }
```

### 2. Проверьте метрики в Prometheus:

- Перейдите в http://localhost:9090/graph
- Введите `orders_created_total` - должно быть 5
- Введите `messages_consumed_total` - должно быть 5
- Введите `payments_processed_total` - увидите распределение success/failed

### 3. Посмотрите dashboard в Grafana:

- http://localhost:3000/d/order-processor
- Все графики обновляются каждые 5 секунд
- Вы увидите рост счетчиков в реальном времени

### 4. Симуляция ошибок:

Платежная система в проекте специально настроена так, что ~10% платежей случайно "фейлятся" для демонстрации retry logic.

После создания 10-20 заказов вы увидите:
- Неудачные платежи в графике "Payment Processing Rate"
- Увеличение счетчика "Messages Retried"
- Возможно появление сообщений в "Messages in DLQ" (если 3 retry не помогли)

## Структура метрик

### order-api метрики:

| Метрика | Тип | Описание |
|---------|-----|----------|
| `orders_created_total` | Counter | Количество созданных заказов |
| `http_request_duration_seconds` | Histogram | Время обработки HTTP запросов |

### payment-worker метрики:

| Метрика | Тип | Описание |
|---------|-----|----------|
| `payments_processed_total{status}` | Counter | Обработанные платежи (success/failed) |
| `messages_consumed_total` | Counter | Потребленные сообщения из Kafka |
| `messages_retried_total` | Counter | Повторные попытки обработки |
| `messages_sent_to_dlq_total` | Counter | Сообщения отправленные в DLQ |
| `payment_processing_duration_seconds` | Histogram | Время обработки платежа |

## Troubleshooting

### Grafana показывает "No data"

1. Проверьте что Prometheus работает: http://localhost:9090/targets
   - Все targets должны быть "UP"
2. Проверьте что datasource настроен:
   - Settings → Data sources → Prometheus
3. Создайте несколько заказов: `.\scripts\test.ps1`

### Prometheus не видит метрики

1. Проверьте endpoints:
   - order-api: http://localhost:8080/metrics
   - payment-worker: http://localhost:8081/metrics
2. Проверьте конфигурацию: `deployments/prometheus.yml`
3. Перезапустите сервисы: `.\scripts\rebuild.ps1`

### Dashboard не появляется

1. Проверьте что volumes смонтированы:
   ```powershell
   docker inspect orders-grafana | findstr grafana
   ```
2. Перезапустите Grafana:
   ```powershell
   docker-compose -f deployments/docker-compose.yml restart grafana
   ```

## Полезные PromQL запросы

```promql
# Процент успешности платежей за последние 5 минут
sum(rate(payments_processed_total{status="success"}[5m])) / sum(rate(payments_processed_total[5m])) * 100

# Средняя скорость создания заказов (в секунду)
rate(orders_created_total[1m])

# Количество ретраев на каждое сообщение
messages_retried_total / messages_consumed_total

# P99 latency обработки платежей
histogram_quantile(0.99, rate(payment_processing_duration_seconds_bucket[5m]))
```
