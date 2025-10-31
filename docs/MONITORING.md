# UAV Heights Map - Руководство по мониторингу

## Обзор

Упрощенная система мониторинга для UAV Heights Map включает:

- **Prometheus** - сбор метрик (только внутренняя Docker сеть)
- **Grafana** - визуализация и дашборды (единый веб-интерфейс)
- **Loki** - централизованное логирование
- **Promtail** - сбор логов
- **Postgres Exporter** - экспорт метрик БД (только основной экспортер)

## Архитектура

```
┌──────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   Приложения     │ --> │   Prometheus    │ --> │    Grafana      │
│                  │     │                 │     │                 │
│ • Go Orchestrator│     │ • Метрики       │     │ • Дашборды      │
│ • Python Svcs    │     │ • Запись        │     │ • Визуализация  │
│ • Frontend       │     │ • Правила       │     │ • Логи и метрики│
│ • Postgres       │     │                 │     │                 │
└──────────────────┘     └─────────────────┘     └─────────────────┘
                                                        ↑
┌─────────────────┐      ┌─────────────────┐     ┌─────────────────┐     
│ Источники логов │  --> │     Promtail    │ --> │      Loki       │
│                 │      │                 │     │                 │
│ • Контейнеры    │      │ • Сбор логов    │     │ • Хранение логов│
│ • Системные логи│      │ • Обработка     │     │ • Запросы       │
│ • Nginx логи    │      │ • Фильтрация    │     │                 │
└─────────────────┘      └─────────────────┘     └─────────────────┘
```

## Компоненты

### Prometheus (Порт: 9090 - Только внутренний)
- **Сбор метрик**: Собирает метрики со всех сервисов и postgres-exporter
- **Хранение**: 30 дней (prod), 7 дней (dev/cicd)
- **Интервал сбора**: 15 секунд для приложений, 30 секунд для инфраструктуры
- **Правила записи**: Предварительно вычисленные метрики для бизнес-анализа и БД
- **Доступ**: Только внутренняя Docker сеть, веб-интерфейс недоступен извне

### Grafana (Порт: 3000)
- **Доступ**:
  - **Production**: http://YOUR_SERVER_IP:3000/
  - **Development**: http://localhost:3000/
- **Логин по умолчанию**: admin/admin (изменить в production)
- **Дашборды**: Автоматически загружаются из `/monitoring/grafana/dashboards/`
- **Возможности**: Просмотр метрик, логов, создание алертов

### Loki (Порт: 3100 - Только внутренний)
- **Хранение логов**: Централизованное хранение логов
- **Хранение**: Настраивается в конфигурации
- **Интеграция**: Интегрирован с Grafana для корреляции метрик и логов
- **Доступ**: Только через Grafana

## Собираемые метрики

### Метрики приложений

#### Go Orchestrator (Порт: 8080)
- `http_requests_total` - общее количество HTTP запросов
- `http_request_duration_seconds` - время ответа
- `heightmap_jobs_total` - одиночные задачи карт высот
- `batch_heightmap_jobs_total` - пакетные задачи
- `heightmap_jobs_completed_total` - завершенные задачи
- `heightmap_job_duration_seconds` - время выполнения
- `rabbitmq_published_tasks_total` - опубликовано задач в RabbitMQ
- `rabbitmq_publish_errors_total` - ошибки публикации в RabbitMQ

#### Python Workers (Порт: 8000)
- `heightmap_single_tasks_total` - одиночные задачи (MiDaS)
- `heightmap_batch_tasks_total` - пакетные задачи (NodeODM)
- `heightmap_processing_duration_seconds` - время обработки
- `heightmap_errors_total` - ошибки генерации
- `heightmap_active_jobs` - активные задачи
- `nodeodm_processing_duration_seconds` - время обработки NodeODM
- `image_resize_operations_total` - операции сжатия изображений (Fast Mode)
- `source_images_deleted_total` - удалено исходных изображений
- `orthophoto_generated_total` - сгенерировано ортофотопланов

#### Frontend (Порт: 80)
- Метрики доступны на эндпоинте `/api/metrics`
- Метрики взаимодействия пользователей и производительности

#### Метрики инфраструктуры (Postgres Exporter)
- `pg_up` - статус PostgreSQL
- `pg_stat_activity_count` - активные соединения с БД
- `pg_database_size_bytes` - размер БД
- `pg_stat_database_*` - статистика БД

## Цели мониторинга

### Задачи сбора Prometheus
- **prometheus** - самомониторинг (localhost:9090)
- **go-orchestrator** - REST API + RabbitMQ publisher (go-orchestrator:8080)
- **python-workers** - MiDaS + NodeODM workers (python-workers:8000)
- **nodeodm** - фотограмметрическая обработка (nodeodm:3000)
- **rabbitmq** - метрики очереди (rabbitmq:15672)
- **frontend** - метрики фронтенда (frontend:80/health)
- **postgres-exporter** - метрики БД (postgres-exporter:9187)
- **grafana** - дашборды (grafana:3000)
- **loki** - логи (loki:3100)

### Источники сбора логов
- **Docker контейнеры** - все логи сервисов через Docker socket
- **Системные логи** - логи хоста через journal
- **Nginx логи** - логи доступа и ошибок

## Дашборды

### 1. Обзор UAV
- Общий статус системы
- Ключевые метрики производительности
- Статус всех сервисов
- Активные алерты

### 2. Метрики Go Orchestrator
- HTTP запросы и время ответа
- Статистика задач карт высот
- Использование ресурсов
- Ошибки аутентификации

### 3. Метрики Python сервисов
- Метрики для каждого Python сервиса
- Время выполнения операций
- Статистика успешности
- Использование ресурсов

### 4. Метрики базы данных
- Соединения и производительность PostgreSQL (через postgres-exporter)
- Размер БД и тренды роста
- Мониторинг активных соединений
- Статус здоровья БД

## Инструкции по настройке

### 1. Переменные окружения
Добавить в файл `.env`:
```env
# Grafana
GRAFANA_ADMIN_USER=admin
GRAFANA_ADMIN_PASSWORD=secure-password
GRAFANA_SERVER_ROOT_URL=https://your-domain.com/grafana

# Мониторинг
POSTGRES_PASSWORD=your-postgres-password
```

### 2. Запуск стека мониторинга
```bash
# Production
docker compose -f deployment/docker/docker-compose.yml up -d

# Development  
docker compose -f deployment/docker/docker-compose.dev.yml up -d
```

### 3. Доступ к мониторингу
- **Grafana**: https://your-domain.com/grafana/
- **Prometheus**: Только внутренняя сеть (недоступен извне)
- **Loki**: Только внутренняя сеть (доступен через Grafana)

### 4. Импорт дашбордов
Дашборды автоматически загружаются из `/monitoring/grafana/dashboards/`

### 5. Проверка настройки
```bash
# Проверить что все контейнеры мониторинга запущены
docker ps | grep -E "(prometheus|grafana|loki|promtail|postgres-exporter)"

# Проверить цели Prometheus
docker exec uav-prometheus curl -s http://localhost:9090/api/v1/targets | jq '.data.activeTargets[].health'

# Проверить источники данных Grafana
curl -s http://localhost:3000/api/datasources | jq '.[].name'
```

## Обслуживание

### Хранение логов
- **Loki**: Настроить хранение в `monitoring/loki/loki.yml`
- **Prometheus**: 30 дней (настраивается в docker-compose)

### Резервное копирование
```bash
# Резервное копирование дашбордов Grafana
docker exec uav-grafana grafana-cli admin export-dashboard

# Резервное копирование данных Prometheus
docker run --rm -v prometheus_data:/data -v $(pwd):/backup alpine tar czf /backup/prometheus-backup.tar.gz /data
```

### Файлы конфигурации
- **Prometheus**: `/monitoring/prometheus/prometheus.yml` - конфигурация сбора и глобальные настройки
- **Loki**: `/monitoring/loki/loki.yml` - конфигурация хранения логов
- **Promtail**: `/monitoring/promtail/promtail.yml` - конфигурация сбора логов
- **Grafana Datasources**: `/monitoring/grafana/provisioning/datasources/datasources.yml`
- **Grafana Dashboards**: `/monitoring/grafana/provisioning/dashboards/dashboards.yml`

## Устранение неполадок

### Частые проблемы

1. **Метрики не появляются**
   - Проверить что сервисы экспортируют метрики на `/metrics`
   - Убедиться что Prometheus может достичь сервисы
   - Проверить конфигурацию `prometheus.yml`

2. **Логи не отображаются в Grafana**
   - Проверить статус Promtail: `docker logs uav-promtail`
   - Убедиться что Loki доступен
   - Проверить конфигурацию `promtail.yml`

3. **Дашборды не загружаются**
   - Проверить логи Grafana: `docker logs uav-grafana`
   - Проверить конфигурацию источников данных в `/monitoring/grafana/provisioning/datasources/`
   - Проверить JSON файлы дашбордов в `/monitoring/grafana/dashboards/`

### Проверки здоровья
```bash
# Проверить цели Prometheus (внутренняя сеть)
docker exec uav-prometheus curl -s http://localhost:9090/api/v1/targets | jq '.data.activeTargets[] | {job: .labels.job, health: .health}'

# Проверить здоровье Loki (внутренняя сеть)
docker exec uav-loki curl http://localhost:3100/ready

# Проверить здоровье Grafana
curl http://localhost:3000/api/health

# Проверить метрики postgres-exporter
docker exec postgres-exporter curl -s http://localhost:9187/metrics | grep pg_up
```

## Соображения безопасности

1. **Контроль доступа**: Использовать надежные пароли для админа Grafana
2. **Сетевая безопасность**: Использовать внутренние сети для компонентов мониторинга
3. **Хранение данных**: Настроить соответствующие политики хранения
4. **Резервное копирование**: Регулярно создавать резервные копии конфигураций и данных