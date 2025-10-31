# UAV Heights Map - Структура проекта

> **Комплексная платформа для обработки и визуализации данных БПЛА**

---

## Обзор проекта

```
uav-heights-map/
├── backend/
│   ├── go-orchestrator/          # Go REST API + RabbitMQ publisher
│   │   ├── cmd/
│   │   │   └── main.go
│   │   ├── config/
│   │   │   └── config.go
│   │   ├── docs/
│   │   │   ├── docs.go
│   │   │   ├── swagger.json
│   │   │   └── swagger.yaml
│   │   ├── internal/
│   │   │   ├── auth/            # Аутентификация и авторизация
│   │   │   │   ├── handler.go
│   │   │   │   ├── interfaces.go
│   │   │   │   ├── payload.go
│   │   │   │   ├── repository.go
│   │   │   │   ├── service.go
│   │   │   │   └── service_test.go
│   │   │   ├── heightmap/       # Обработка карт высот
│   │   │   │   ├── handler.go
│   │   │   │   ├── interfaces.go
│   │   │   │   ├── model.go
│   │   │   │   ├── payload.go
│   │   │   │   ├── service.go
│   │   │   │   └── service_test.go
│   │   │   └── storage/         # Работа с БД и MinIO
│   │   │       ├── db.go
│   │   │       ├── schema.sql
│   │   │       ├── migrations/
│   │   │       │   ├── 000001_initial_schema.up.sql
│   │   │       │   ├── 000001_initial_schema.down.sql
│   │   │       │   ├── 000002_create_heightmap_jobs.up.sql
│   │   │       │   └── 000002_create_heightmap_jobs.down.sql
│   │   │       ├── queries/
│   │   │       │   ├── users.sql
│   │   │       │   └── heightmap.sql
│   │   │       ├── sqlc/
│   │   │       │   ├── db.go
│   │   │       │   ├── models.go
│   │   │       │   ├── users.sql.go
│   │   │       │   └── heightmap.sql.go
│   │   │       └── minio/
│   │   │           └── minio.go
│   │   ├── pkg/                 # Общие пакеты
│   │   │   ├── jwt/
│   │   │   │   └── jwt.go
│   │   │   ├── metrics/
│   │   │   │   └── metrics.go
│   │   │   ├── middleware/
│   │   │   │   ├── auth.go
│   │   │   │   └── logger.go
│   │   │   └── rabbitmq/
│   │   │       ├── client.go
│   │   │       └── messages.go
│   │   ├── sqlc.yaml
│   │   ├── go.mod
│   │   ├── go.sum
│   │   ├── Dockerfile
│   │   └── Makefile
│   │
│   └── python-services/
│       └── heightmap-service/   # Python Workers (RabbitMQ)
│           ├── heightmap_service/
│           │   ├── __init__.py
│           │   ├── config.py
│           │   ├── core/
│           │   │   ├── logging.py
│           │   │   ├── metrics.py
│           │   │   └── metrics_server.py
│           │   ├── workers/
│           │   │   ├── single_worker.py     # MiDaS обработка
│           │   │   └── batch_worker.py      # NodeODM обработка
│           │   ├── services/
│           │   │   ├── image_service.py     # MiDaS
│           │   │   └── batch_service.py     # NodeODM клиент
│           │   ├── infrastructure/
│           │   │   └── minio_client.py
│           │   └── utils/
│           │       └── exif_helper.py
│           ├── requirements.txt
│           └── Dockerfile
│
├── deployment/                 # Конфигурация развертывания
│   ├── docker/
│   │   ├── docker-compose.yml      # Production
│   │   ├── docker-compose.dev.yml  # Development
│   │   └── docker-compose.cicd.yml # CI/CD
│   ├── nginx/
│   │   ├── prod.conf
│   │   └── dev.conf
│   ├── scripts/
│   │   └── rollback.sh
│   └── Makefile
│
├── frontend/                   # Next.js фронтенд
│   ├── src/
│   │   ├── app/
│   │   │   ├── layout.tsx
│   │   │   ├── page.tsx
│   │   │   ├── globals.css
│   │   │   ├── api/
│   │   │   │   ├── metrics/
│   │   │   │   └── health/
│   │   │   ├── projects/
│   │   │   └── processing/
│   │   ├── components/
│   │   │   ├── ui/
│   │   │   │   └── Button.tsx
│   │   │   ├── map/
│   │   │   │   └── MapViewer.tsx
│   │   │   └── upload/
│   │   │       └── FileUpload.tsx
│   │   ├── services/
│   │   │   ├── api.ts
│   │   │   ├── projectService.ts
│   │   │   └── processingService.ts
│   │   ├── utils/
│   │   │   ├── format.ts
│   │   │   └── validation.ts
│   │   └── middleware.ts
│   ├── public/
│   ├── package.json
│   ├── next.config.js
│   ├── tsconfig.json
│   ├── tailwind.config.js
│   ├── postcss.config.js
│   ├── .eslintrc.json
│   ├── .gitignore
│   ├── .dockerignore
│   └── Dockerfile
│
├── monitoring/                 # Мониторинг и метрики
│   ├── grafana/
│   │   ├── dashboards/
│   │   │   ├── metrics/
│   │   │   │   └── system-overview.json
│   │   │   ├── logs/
│   │   │   │   └── system-logs.json
│   │   │   ├── go-orchestrator/
│   │   │   │   ├── metrics/
│   │   │   │   │   └── go-orchestrator-overview.json
│   │   │   │   └── logs/
│   │   │   │       └── go-orchestrator-logs.json
│   │   │   ├── python-services/
│   │   │   │   ├── metrics/
│   │   │   │   │   └── python-services-overview.json
│   │   │   │   └── logs/
│   │   │   │       └── python-services-logs.json
│   │   │   ├── frontend/
│   │   │   │   ├── metrics/
│   │   │   │   │   └── frontend-overview.json
│   │   │   │   └── logs/
│   │   │   │       └── frontend-logs.json
│   │   │   ├── nginx/
│   │   │   │   ├── metrics/
│   │   │   │   │   └── nginx-overview.json
│   │   │   │   └── logs/
│   │   │   │       └── nginx-logs.json
│   │   │   ├── postgresql/
│   │   │   │   ├── metrics/
│   │   │   │   │   └── postgresql-overview.json
│   │   │   │   └── logs/
│   │   │   │       └── postgresql-logs.json
│   │   │   └── additional/
│   │   │       ├── metrics/
│   │   │       │   └── monitoring-services.json
│   │   │       └── logs/
│   │   │           └── monitoring-services-logs.json
│   │   └── provisioning/
│   │       ├── dashboards/
│   │       │   └── dashboards.yml
│   │       └── datasources/
│   │           └── datasources.yml
│   ├── prometheus/
│   │   ├── prometheus.yml
│   │   └── rules/
│   │       └── recording.yml
│   ├── loki/
│   │   └── loki.yml
│   └── promtail/
│       └── promtail.yml
│
├── docs/                       # Документация
│   ├── PROJECT_STRUCTURE.md
│   ├── ARCHITECTURE.md
│   ├── API-SPEC.md
│   ├── DATABASE.md
│   └── MONITORING.md
│
├── .gitlab-ci.yml              # GitLab CI/CD конфигурация
├── .gitlab-ci-prod.yml
├── .gitlab-ci-dev.yml
└── README.md
```

---

## Основные компоненты архитектуры

### **Backend сервисы**
- **Go Orchestrator**: REST API с JWT, RabbitMQ publisher, управление задачами
- **Python Workers**: Асинхронная обработка через RabbitMQ (MiDaS для одиночных, NodeODM для пакетных)
- **NodeODM**: Фотограмметрическая обработка (DSM + ортофотопланы)
- **Инфраструктура**: PostgreSQL (SQLC), MinIO S3 (3 бакета), RabbitMQ (очереди)

### **Frontend приложение**
- **Стек**: React 18 + Next.js 14 + TypeScript + Tailwind CSS
- **UI**: Темная/светлая тема, адаптивный дизайн
- **Обработка**: Выбор качества (Ультра/Среднее/Низкое), Fast Mode, тип продукта (DSM/Ортофотоплан)
- **Аутентификация**: Автообновление токенов при 401

### **Ключевые возможности**
- **Одиночная генерация**: MiDaS + CLAHE + Edge Enhancement + Hillshade (~2-5 сек)
- **Пакетная генерация**: NodeODM для фотограмметрии (DSM + Ортофотопланы, требует ≥5 фото с GPS)
- **3 уровня качества**: Ультра/Среднее/Низкое с настройками NodeODM
- **Fast Mode**: Уменьшение изображений до 2000px + флаги оптимизации (ускорение до 75%)
- **Автоудаление**: Исходные фото удаляются после успешной обработки
- **RabbitMQ**: Асинхронная очередь с автореконнектом
- **Публичные бакеты**: 3 MinIO бакета (`uav-data`, `uav-models`, `uav-photoplanes`)
- **JWT + Auto-refresh**: Автообновление токенов при истечении (401)

### **Архитектура базы данных**
- **PostgreSQL**: Основная реляционная БД с поддержкой JSONB
- **Паттерн Repository**: Чистое разделение между доступом к данным и бизнес-логикой
- **Миграции БД**: golang-migrate для версионирования схемы и развертывания
- **Типобезопасные запросы**: SQLC для генерации типобезопасного Go кода из SQL
- **Пул соединений**: Оптимизированное управление соединениями с БД
- **Мониторинг здоровья**: Встроенные проверки здоровья БД и статистика

---

## Быстрый старт

```bash
# Настройка окружения
cd deployment
make setup

# Запуск всего стека (включая PostgreSQL, MinIO, мониторинг)
make up

# Запуск миграций БД
cd ../backend/go-orchestrator
make migrate

# Запуск фронтенд dev сервера
cd ../../frontend
npm install
npm run dev
```

### **Команды разработки Backend**
```bash
# Запуск миграций
cd backend/go-orchestrator && make migrate

# Откат миграций
make migrate-down

# Генерация типобезопасного Go кода из SQL запросов
make sqlc

# Сборка Go оркестратора
make build

# Запуск тестов
make test
```

### **Команды Docker развертывания**
```bash
# Production развертывание
cd deployment
make up

# Development окружение
make dev

# Проверка статуса сервисов
make status

# Просмотр логов
make logs
```

---

## Порты сервисов

### **Production окружение**
- **nginx**: `81` (HTTP)
- **go-orchestrator**: `8080` (REST API)
- **python-workers**: `8000` (Metrics/Health)
- **nodeodm**: `3000` (HTTP API)
- **postgresql**: `5432`
- **rabbitmq**: `5672` (AMQP), `15672` (Management UI)
- **minio**: `9000` (S3 API), `9001` (Console)
- **prometheus**: `9090`
- **grafana**: `3000` (дашборды)
- **loki**: `3100`

### **Development окружение**
- Все сервисы с префиксом `dev-uav-*`
- Открытые порты такие же как в production для локального доступа

---

## Технологический стек

### **Backend**
- **Go**: 1.25+ (API Orchestrator)
- **Python**: 3.11+ (Workers: MiDaS, NodeODM client, Pillow, OpenCV, GDAL)
- **PostgreSQL**: 17+ (SQLC для Go)
- **MinIO**: S3-совместимое хранилище (3 публичных бакета)
- **RabbitMQ**: Очередь сообщений (автореконнект)
- **NodeODM**: OpenDroneMap для фотограмметрии

### **Frontend**
- **Next.js**: 14+ (React фреймворк)
- **TypeScript**: 5+
- **Tailwind CSS**: 3+ (Стилизация + темная/светлая тема)

### **Мониторинг и наблюдаемость**
- **Prometheus**: Сбор метрик
- **Grafana**: Дашборды и визуализация
- **Loki**: Агрегация логов
- **Promtail**: Доставка логов

### **CI/CD**
- **GitLab CI/CD**: Автоматизированные сборки и развертывания
- **Docker**: Контейнеризация
- **Docker Compose**: Оркестрация

---

*Создано для обработки и визуализации данных БПЛА*