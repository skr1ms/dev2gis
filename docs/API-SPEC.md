# Спецификация API

## Базовый URL
- **Development**: `http://localhost/api`
- **Production**: `http://YOUR_SERVER_IP:81/api`

## Интерактивная документация
- **Swagger UI (dev)**: `http://localhost/swagger/index.html`
- **Swagger UI (prod)**: `http://YOUR_SERVER_IP:81/swagger/index.html`
- **Автогенерируемая**: Используя goswag + swag интеграцию
- **Без аутентификации**: Swagger документация публично доступна

## Основная концепция

Этот API предназначен для **генерации карт высот (DSM) и ортофотопланов из фотографий БПЛА**:

1. **Загрузить 1 фото** → **Сгенерировать карту высот (MiDaS)**
2. **Загрузить 2+ фото с GPS** → **Сгенерировать DSM и/или ортофотоплан (NodeODM)**
3. **Выбрать качество и режим** → **Оптимизировать скорость или качество**

## Аутентификация

Все защищенные эндпоинты требуют JWT аутентификации через заголовок `Authorization`:
```
Authorization: Bearer <jwt_token>
```

## Эндпоинты

### Аутентификация

#### POST /api/auth/register
Регистрация нового пользователя.

**Тело запроса:**
```json
{
  "email": "string",
  "password": "string",
  "name": "string"
}
```

**Ответ:**
```json
{
  "user": {
    "id": "uuid",
    "email": "string",
    "name": "string",
    "created_at": "timestamp"
  },
  "tokens": {
    "access_token": "string",
    "refresh_token": "string"
  }
}
```

#### POST /api/auth/login
Аутентификация пользователя и получение JWT токенов.

**Тело запроса:**
```json
{
  "email": "string",
  "password": "string"
}
```

**Ответ:**
```json
{
  "user": {
    "id": "uuid",
    "email": "string",
    "name": "string"
  },
  "tokens": {
    "access_token": "string",
    "refresh_token": "string"
  }
}
```

#### POST /api/auth/refresh
Обновление access токена используя refresh токен. **Автоматически** вызывается фронтендом при истечении токена (401).

**Тело запроса:**
```json
{
  "refresh_token": "string"
}
```

**Ответ:**
```json
{
  "access_token": "string",
  "refresh_token": "string"
}
```

---

### Карты высот

#### POST /api/heightmaps/upload
🔒 **Требуется аутентификация** - Загрузить **одно** изображение БПЛА и создать задачу генерации карты высот.

**Запрос:** `multipart/form-data`
```
file: binary (обязательно)
```

**Ответ:**
```json
{
  "id": "uuid",
  "status": "pending"
}
```

#### POST /api/heightmaps/batch/upload
🔒 **Требуется аутентификация** - Загрузить **несколько** изображений БПЛА для пакетной генерации карты высот и/или ортофотоплана.

**Запрос:** `multipart/form-data`
```
files[]: binary[] (обязательно, мин. 2 изображения)
merge_method: string (опц., "max"|"average"|"low", по умолчанию "average")
fast_mode: boolean (опц., true для быстрой генерации с уменьшением изображений, по умолчанию false)
generation_mode: string (опц., "heightmap"|"orthophoto"|"both", по умолчанию "heightmap")
```

**Параметры:**
- `merge_method`: Качество обработки
  - `"max"` - Ультра (макс. качество, медленно)
  - `"average"` - Среднее (баланс скорости и качества)
  - `"low"` - Низкое (быстро, базовое качество)
- `fast_mode`: При `true` изображения уменьшаются до 2000px и используются флаги оптимизации NodeODM
- `generation_mode`: 
  - `"heightmap"` - только карта высот (DSM)
  - `"orthophoto"` - только ортофотоплан (требует ≥5 фото)
  - `"both"` - оба продукта (требует ≥5 фото)

**Ответ:**
```json
{
  "batch_job_id": "uuid",
  "status": "pending",
  "image_count": 10,
  "merge_method": "average",
  "generation_mode": "heightmap"
}
```

#### GET /api/heightmaps/batch/:id
🔒 **Требуется аутентификация** - Получить статус пакетной задачи.

**Ответ:**
```json
{
  "id": "uuid",
  "user_id": "uuid",
  "status": "pending|processing|completed|failed",
  "result_url": "string",
  "orthophoto_url": "string",
  "image_count": 10,
  "processed_count": 8,
  "merge_method": "average",
  "generation_mode": "heightmap",
  "processing_time": 45.2,
  "created_at": "timestamp"
}
```

#### GET /api/heightmaps/:id
🔒 **Требуется аутентификация** - Получить детали задачи карты высот по ID.

**Параметры:**
- `id` (path): UUID задачи карты высот

**Ответ:**
```json
{
  "id": "uuid",
  "user_id": "uuid",
  "image_url": "string",
  "result_url": "string",
  "status": "pending|processing|completed|failed",
  "width": 1024,
  "height": 768,
  "error_message": "string",
  "processing_time": 12.5,
  "created_at": "timestamp",
  "updated_at": "timestamp"
}
```

#### GET /api/heightmaps
🔒 **Требуется аутентификация** - Получить список всех задач карт высот для аутентифицированного пользователя.

**Параметры запроса:**
- `limit` (опционально): Количество элементов на странице (по умолчанию: 20, максимум: 100)
- `offset` (опционально): Смещение для пагинации (по умолчанию: 0)

**Ответ:**
```json
{
  "heightmaps": [
    {
      "id": "uuid",
      "user_id": "uuid",
      "image_url": "string",
      "result_url": "string",
      "status": "pending|processing|completed|failed",
      "width": 1024,
      "height": 768,
      "error_message": "string",
      "processing_time": 12.5,
      "created_at": "timestamp",
      "updated_at": "timestamp"
    }
  ],
  "limit": 20,
  "offset": 0
}
```

### Проверка здоровья

#### GET /health
Проверить статус здоровья API сервера.

**Ответ:**
```json
{
  "status": "healthy",
  "environment": "dev|production"
}
```

---

## Поток использования

### 1. Поток аутентификации
```
POST /api/auth/register → Получить токены → Доступ к защищенным эндпоинтам
POST /api/auth/login    → Получить токены → Доступ к защищенным эндпоинтам
POST /api/auth/refresh  → Обновить токены при необходимости
```

### 2. Поток карты высот
```
POST /api/heightmaps/upload → Загрузить фото
GET /api/heightmaps/:id     → Проверить статус обработки
GET /api/heightmaps         → Получить список всех карт высот
```

## Коды статуса ответов

- **200 OK** - Успех
- **201 Created** - Ресурс успешно создан
- **202 Accepted** - Запрос принят к обработке
- **400 Bad Request** - Некорректные данные запроса
- **401 Unauthorized** - Требуется аутентификация или неверный токен
- **404 Not Found** - Ресурс не найден
- **500 Internal Server Error** - Ошибка сервера

## Формат ответа об ошибке
```json
{
  "error": "Описание ошибки"
}
```

## Техническая реализация

### JWT токены
- **Access Token**: 30 минут TTL
- **Refresh Token**: 7 дней TTL
- **Хранение**: Клиентская сторона (localStorage/cookies)

### Загрузка файлов
- **Максимальный размер**: Настраивается (по умолчанию: 100MB)
- **Форматы**: JPEG, PNG
- **Хранение**: MinIO S3

### Обработка карт высот

**Одиночная обработка:**
- **Технология**: MiDaS depth estimation + CLAHE + Edge Enhancement + Multi-angle Hillshade
- **Время**: ~2-5 сек/фото
- **Автоудаление**: Исходное фото удаляется после успешной генерации

**Пакетная обработка:**
- **Технология**: NodeODM (OpenDroneMap)
- **Очередь**: RabbitMQ для асинхронной обработки
- **Качество**:
  - `max` (Ультра): `pc-quality: ultra`, `feature-quality: ultra`, высокое разрешение DEM/Orthophoto
  - `average` (Среднее): `pc-quality: high`, `feature-quality: high`, среднее разрешение
  - `low` (Низкое): `pc-quality: medium`, `feature-quality: high`, низкое разрешение
- **Fast Mode**: 
  - Уменьшение изображений до 2000px (сохранение EXIF/GPS)
  - Флаги NodeODM: `fast-orthophoto`, `skip-3dmodel`, `skip-report`, `optimize-disk-space`, `pc-skip-geometric`
  - Ускорение: ~50-75%
- **Продукты**:
  - **DSM** (Digital Surface Model) - карта высот
  - **Ортофотоплан** - геопривязанное изображение (требует ≥5 фото с GPS)
- **Автоудаление**: Исходные фото удаляются после успешной генерации

### База данных
- **Технология**: PostgreSQL с SQLC генерируемыми запросами
- **Миграции**: golang-migrate для управления схемой
- **Соединения**: Пул соединений с pgxpool