# Схема базы данных

## Таблицы

### users (Пользователи)
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

### heightmap_jobs (Одиночные задачи карт высот)
```sql
CREATE TABLE heightmap_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    image_url VARCHAR(500) NOT NULL,
    result_url VARCHAR(500),
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    width INTEGER,
    height INTEGER,
    error_message TEXT,
    processing_time FLOAT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

### batch_heightmap_jobs (Пакетные задачи)
```sql
CREATE TABLE batch_heightmap_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    result_url TEXT,                          -- URL карты высот (DSM)
    orthophoto_url TEXT,                      -- URL ортофотоплана
    width INTEGER,
    height INTEGER,
    image_count INTEGER NOT NULL DEFAULT 0,   -- Всего изображений
    processed_count INTEGER NOT NULL DEFAULT 0, -- Обработано изображений
    error_message TEXT,
    processing_time REAL,
    merge_method VARCHAR(50) NOT NULL DEFAULT 'average', -- max|average|low
    generation_mode VARCHAR(50) NOT NULL DEFAULT 'heightmap', -- heightmap|orthophoto|both
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

## Индексы

```sql
-- Индексы для одиночных задач
CREATE INDEX idx_heightmap_jobs_user_id ON heightmap_jobs(user_id);
CREATE INDEX idx_heightmap_jobs_status ON heightmap_jobs(status);
CREATE INDEX idx_heightmap_jobs_created_at ON heightmap_jobs(created_at DESC);

-- Индексы для пакетных задач
CREATE INDEX idx_batch_heightmap_jobs_user_id ON batch_heightmap_jobs(user_id);
CREATE INDEX idx_batch_heightmap_jobs_status ON batch_heightmap_jobs(status);
CREATE INDEX idx_batch_heightmap_jobs_created_at ON batch_heightmap_jobs(created_at DESC);

-- Индексы для пользователей
CREATE INDEX idx_users_email ON users(email);
```

## Связи

- `users` → `heightmap_jobs` (1:N) - Пользователь может иметь множество одиночных задач
- `users` → `batch_heightmap_jobs` (1:N) - Пользователь может иметь множество пакетных задач

## Соображения безопасности

### Хеширование паролей
- Пароли хешируются с использованием bcrypt с солью
- Минимальные требования к сложности пароля
- Блокировка аккаунта после неудачных попыток

### JWT токены
- Access токены имеют короткий срок жизни (30 минут)
- Refresh токены имеют длительный срок жизни (7 дней)
- Токены хранятся только на клиентской стороне

## Миграции

### 000001_initial_schema.up.sql
```sql
-- Создание таблицы пользователей
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Создание индексов для пользователей
CREATE INDEX idx_users_email ON users(email);
```

### 000002_create_heightmap_jobs.up.sql
```sql
-- Создание таблицы задач карт высот
CREATE TABLE heightmap_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    image_url VARCHAR(500) NOT NULL,
    result_url VARCHAR(500),
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    width INTEGER,
    height INTEGER,
    error_message TEXT,
    processing_time FLOAT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Создание индексов для задач карт высот
CREATE INDEX idx_heightmap_jobs_user_id ON heightmap_jobs(user_id);
CREATE INDEX idx_heightmap_jobs_status ON heightmap_jobs(status);
CREATE INDEX idx_heightmap_jobs_created_at ON heightmap_jobs(created_at DESC);
```

## SQLC запросы

### users.sql
```sql
-- name: Create :one
INSERT INTO users (email, password_hash, name)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: Delete :exec
DELETE FROM users WHERE id = $1;
```

### heightmap.sql
```sql
-- name: CreateHeightmapJob :one
INSERT INTO heightmap_jobs (user_id, image_url, status)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetHeightmapJob :one
SELECT * FROM heightmap_jobs WHERE id = $1;

-- name: GetHeightmapJobByUserID :one
SELECT * FROM heightmap_jobs WHERE id = $1 AND user_id = $2;

-- name: UpdateJobStatus :exec
UPDATE heightmap_jobs SET status = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $1;

-- name: UpdateJobResult :exec
UPDATE heightmap_jobs 
SET result_url = $2, status = $3, width = $4, height = $5, processing_time = $6, updated_at = CURRENT_TIMESTAMP 
WHERE id = $1;

-- name: UpdateJobError :exec
UPDATE heightmap_jobs 
SET status = $2, error_message = $3, updated_at = CURRENT_TIMESTAMP 
WHERE id = $1;

-- name: ListUserHeightmaps :many
SELECT * FROM heightmap_jobs 
WHERE user_id = $1 
ORDER BY created_at DESC 
LIMIT $2 OFFSET $3;

-- name: CountUserHeightmaps :one
SELECT COUNT(*) FROM heightmap_jobs WHERE user_id = $1;

-- name: ListHeightmapsByStatus :many
SELECT * FROM heightmap_jobs WHERE status = $1 ORDER BY created_at DESC;

-- name: DeleteHeightmapJob :exec
DELETE FROM heightmap_jobs WHERE id = $1 AND user_id = $2;
```

## Конфигурация подключения

### Переменные окружения
```bash
# База данных
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_USER=root
POSTGRES_PASSWORD=your_password
POSTGRES_DB=uav_heights_map

# URL подключения
DATABASE_URL=postgres://root:your_password@postgres:5432/uav_heights_map?sslmode=disable
```

### Пул соединений
- **Максимальные соединения**: 25
- **Минимальные соединения**: 5
- **Таймаут подключения**: 30 секунд
- **Таймаут запроса**: 10 секунд

## Мониторинг

### Метрики PostgreSQL
- Количество активных соединений
- Время выполнения запросов
- Размер базы данных
- Статистика по таблицам

### Проверка здоровья
- Проверка подключения к БД
- Проверка доступности таблиц
- Проверка производительности запросов