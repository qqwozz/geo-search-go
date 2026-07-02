# Geo Search — B2C Умный Геопоисковик

Платформа для поиска мест (кафе, ресторанов, баров) на естественном русском языке. Пользователь пишет запрос типа «тихое кафе с террасой рядом с метро», система находит подходящие места с объяснениями.

## Архитектура

```
┌─────────────┐     ┌──────────────┐     ┌──────────────┐
│   Frontend   │────▶│  Go Backend  │────▶│  Python NLP  │
│  (nginx)     │     │   (Gin)      │     │  (FastAPI)   │
└─────────────┘     └──────┬───────┘     └──────────────┘
                           │
                    ┌──────┴───────┐
                    │              │
              ┌─────▼────┐  ┌─────▼────┐
              │PostgreSQL │  │  Redis    │
              │ +PostGIS  │  │  Cache   │
              └──────────┘  └──────────┘
```

| Сервис | Порт | Стек | Описание |
|--------|------|------|----------|
| **api** | 8080 | Go 1.25, Gin, pgx/v5 | Основной бэкенд, REST API |
| **nlp** | 8000 | Python 3.12, FastAPI | Парсинг текста в фильтры |
| **postgres** | 5432 | PostgreSQL 16 + PostGIS 3.4 | Хранение POI, пространственные запросы |
| **redis** | 6379 | Redis 7 | Кеширование результатов |
| **frontend** | 3000 | nginx | Статический frontend (заглушка) |

## Быстрый старт

### Запуск через Docker

```bash
git clone <repo-url>
cd geo-search-go
docker compose up -d
```

Сервисы:
- API: `http://localhost:8080`
- NLP: `http://localhost:8000`
- Postgres: `localhost:5432`
- Redis: `localhost:6379`

### Проверка работоспособности

```bash
# Health check
curl http://localhost:8080/api/health

# Поиск мест
curl "http://localhost:8080/api/search?q=тихое+кафе+с+террасой&lat=55.755&lon=37.605&radius=2000"

# NLP парсинг
curl -X POST http://localhost:8000/parse \
  -H "Content-Type: application/json" \
  -d '{"text":"тихое кафе с террасой"}'
```

## API Reference

### GET /api/search

Поиск мест по естественному запросу.

**Query Parameters:**

| Параметр | Тип | Обязательный | По умолчанию | Описание |
|----------|-----|--------------|--------------|----------|
| `q` | string | Да | — | Текстовый запрос на русском языке (макс. 500 символов) |
| `lat` | float | Да | — | Широта центра поиска |
| `lon` | float | Да | — | Долгота центра поиска |
| `radius` | int | Нет | 2000 | Радиус поиска в метрах |
| `limit` | int | Нет | 20 | Количество результатов (макс. 50) |
| `sort` | string | Нет | relevance | Сортировка: `relevance`, `distance`, `rating` |

**Пример запроса:**

```bash
curl "http://localhost:8080/api/search?q=уютное+кафе+где+можно+поработать&lat=55.755&lon=37.605&radius=2000"
```

**Пример ответа:**

```json
{
  "pois": [
    {
      "id": 3,
      "name": "Тихий Дворик",
      "category": "cafe",
      "address": "пер. Малый Казённый 7",
      "city": "Москва",
      "rating": 4.7,
      "review_count": 95,
      "price_level": 1,
      "features": {
        "wifi": true,
        "outlets": true,
        "terrace": false,
        "parking": false,
        "live_music": false,
        "breakfast": false,
        "quiet": true,
        "family_friendly": false,
        "romantic": false,
        "dog_friendly": false
      },
      "noise_level": "medium",
      "lat": 55.76,
      "lon": 37.6116,
      "distance_meters": 414.34,
      "explanation": "Отличное место для работы: быстрый Wi-Fi, удобные розетки, очень тихо",
      "score": 10.12
    }
  ],
  "total": 4,
  "center": { "lat": 55.755, "lon": 37.605 },
  "query": "уютное кафе где можно поработать",
  "cached": false
}
```

**Поля ответа:**

| Поле | Описание |
|------|----------|
| `pois[]` | Массив найденных мест |
| `pois[].score` | Балл ранжирования (чем выше — тем лучше подходит) |
| `pois[].explanation` | Человеческое объяснение на русском |
| `pois[].distance_meters` | Расстояние от центра поиска в метрах |
| `pois[].features` | Доступные удобства места |
| `total` | Количество результатов |
| `cached` | Взят ли результат из кеша |

### GET /api/autocomplete

Подсказки при вводе запроса.

**Query Parameters:**

| Параметр | Тип | Обязательный | Описание |
|----------|-----|--------------|----------|
| `q` | string | Нет | Начало запроса для фильтрации |

**Пример:**

```bash
curl "http://localhost:8080/api/autocomplete?q=кафе"
```

**Ответ:**

```json
{
  "suggestions": [
    "кафе с террасой",
    "тихое кафе с вайфаем",
    "кафе с розетками",
    "кафе для работы",
    "кафе с парковкой"
  ]
}
```

### GET /api/health

Проверка работоспособности всех сервисов.

**Ответ (200 OK):**

```json
{
  "status": "ok",
  "postgres": "ok",
  "redis": "ok"
}
```

### POST /parse (NLP сервис, порт 8000)

Парсинг текстового запроса в структурированные фильтры.

**Request Body:**

```json
{
  "text": "тихое кафе с террасой рядом с метро Парк Культуры",
  "city": "moscow"
}
```

**Ответ:**

```json
{
  "category": "cafe",
  "intent": "default",
  "features": {
    "terrace": true,
    "quiet": true
  },
  "location": {
    "metro": "Парк Культуры"
  },
  "radius": 500,
  "radius_raw": "рядом",
  "sort_by": "relevance"
}
```

**Поля NLP ответа:**

| Поле | Описание |
|------|----------|
| `category` | Категория: `cafe`, `restaurant`, `bar`, `fast_food`, `park` |
| `intent` | Намерение: `work`, `breakfast`, `dinner`, `romantic`, `default` |
| `features` | Обнаруженные фичи: wifi, terrace, quiet, outlets, breakfast, parking, romantic, family, live_music |
| `location` | Локация: metro, street, area |
| `radius` | Радиус поиска в метрах (из текста) |

## Структура проекта

```
geo-search-go/
├── cmd/
│   └── api/main.go                  # Точка входа Go API
├── internal/
│   ├── config/config.go             # Конфигурация из env-переменных
│   ├── database/
│   │   ├── postgres.go              # Инициализация pgx pool
│   │   └── redis.go                 # Инициализация Redis клиента
│   ├── handlers/
│   │   ├── search.go                # GET /api/search
│   │   ├── autocomplete.go          # GET /api/autocomplete
│   │   └── health.go                # GET /api/health
│   ├── models/poi.go                # Структуры данных (POI, SearchRequest, NLPResponse)
│   └── services/
│       ├── search.go                # Оркестратор поиска
│       ├── nlp.go                   # HTTP клиент к Python NLP
│       ├── cache.go                 # Redis кеширование
│       ├── fallback_parser.go       # Regex парсер (fallback при недоступности NLP)
│       ├── ranker.go                # Intent-based ранжирование
│       └── explainer.go             # Генерация объяснений
├── nlp/
│   ├── main.py                      # FastAPI приложение
│   ├── parser.py                    # Логика парсинга
│   ├── dictionaries.py              # Словари русских слов
│   ├── location_parser.py           # Извлечение метро/улиц
│   ├── requirements.txt             # Python зависимости
│   ├── Dockerfile
│   └── tests/
│       ├── test_parser.py
│       └── test_location.py
├── scripts/
│   ├── ingest_osm.py                # Импорт данных из OpenStreetMap
│   └── ingest_config.json           # Конфигурация импорта
├── migrations/
│   ├── 001_create_database.sql      # Расширение PostGIS
│   ├── 002_create_pois_table.sql    # Таблица POI
│   └── 003_create_poi_indexes.sql   # Индексы
├── docker-compose.yml
├── Dockerfile                       # Go контейнер
├── .env.example
├── go.mod
└── go.sum
```

## Схема базы данных

### Таблица `pois`

```sql
CREATE TABLE pois (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    name_en VARCHAR(255),
    category VARCHAR(50) NOT NULL,        -- cafe, restaurant, bar, fast_food, park
    subcategory VARCHAR(50),
    address TEXT,
    city VARCHAR(100),
    phone VARCHAR(20),
    website VARCHAR(500),
    opening_hours JSONB,
    rating DECIMAL(2,1) DEFAULT 0,        -- 0.0 - 5.0
    review_count INTEGER DEFAULT 0,
    price_level INTEGER DEFAULT 2,        -- 1=budget, 2=moderate, 3=expensive
    has_wifi BOOLEAN DEFAULT FALSE,
    has_outlets BOOLEAN DEFAULT FALSE,
    has_terrace BOOLEAN DEFAULT FALSE,
    has_parking BOOLEAN DEFAULT FALSE,
    has_live_music BOOLEAN DEFAULT FALSE,
    has_breakfast BOOLEAN DEFAULT FALSE,
    is_quiet BOOLEAN DEFAULT FALSE,
    is_family_friendly BOOLEAN DEFAULT FALSE,
    is_romantic BOOLEAN DEFAULT FALSE,
    is_dog_friendly BOOLEAN DEFAULT FALSE,
    noise_level VARCHAR(10) DEFAULT 'medium',
    osm_id BIGINT UNIQUE,
    osm_type VARCHAR(10),
    source VARCHAR(50) DEFAULT 'osm',
    geom GEOGRAPHY(POINT, 4326) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    last_updated TIMESTAMP DEFAULT NOW()
);
```

**Индексы:**

```sql
CREATE INDEX idx_pois_geom ON pois USING GIST (geom);
CREATE INDEX idx_pois_category ON pois (category);
CREATE INDEX idx_pois_rating ON pois (rating DESC);
CREATE INDEX idx_pois_city ON pois (city);
CREATE INDEX idx_pois_wifi ON pois (has_wifi) WHERE has_wifi = TRUE;
CREATE INDEX idx_pois_terrace ON pois (has_terrace) WHERE has_terrace = TRUE;
CREATE INDEX idx_pois_quiet ON pois (is_quiet) WHERE is_quiet = TRUE;
CREATE INDEX idx_pois_breakfast ON pois (has_breakfast) WHERE has_breakfast = TRUE;
```

## Алгоритм ранжирования

Ранжирование зависит от намерения (intent), определённого NLP:

### Work (работа)

```
score = (wifi × 3) + (outlets × 3) + (quiet × 2) + (rating/5 × 2) + (1/distance × 100) - (live_music × 3)
```

### Breakfast (завтрак)

```
score = (breakfast × 5) + (rating/5 × 3) + (price_level ≤ 2 × 2) + (1/distance × 50)
```

### Romantic (романтика)

```
score = (romantic × 4) + (rating/5 × 2) + (quiet × 2) + (1/distance × 30)
```

### Default (общий)

```
score = (rating/5 × 5) + (1/distance × 50) + (review_count/100 × 2)
```

## Конфигурация

### Env-переменные

| Переменная | По умолчанию | Описание |
|------------|--------------|----------|
| `PORT` | 8080 | Порт Go API |
| `DATABASE_URL` | postgres://postgres:postgres@localhost:5432/geosearch | URL подключения к PostgreSQL |
| `REDIS_URL` | redis://localhost:6379 | URL подключения к Redis |
| `NLP_SERVICE_URL` | http://localhost:8000 | URL Python NLP сервиса |
| `CORS_ORIGIN` | http://localhost:3000 | Разрешённый_ORIGIN для CORS |

### Создание .env файла

```bash
cp .env.example .env
# Отредактируй под свою среду
```

## Разработка

### Запуск локально (без Docker)

```bash
# 1. PostgreSQL + PostGIS
# Должен быть запущен на localhost:5432

# 2. Redis
# Должен быть запущен на localhost:6379

# 3. Go API
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/geosearch?sslmode=disable"
export REDIS_URL="redis://localhost:6379"
export NLP_SERVICE_URL="http://localhost:8000"
go run cmd/api/main.go

# 4. Python NLP
cd nlp
pip install -r requirements.txt
uvicorn main:app --host 0.0.0.0 --port 8000
```

### Тесты

```bash
# Go тесты
go test ./internal/services/... -v

# Python тесты
cd nlp && pytest tests/ -v
```

### Импорт данных из OpenStreetMap

```bash
# Установка зависимостей
pip install overpy psycopg2-binary

# Импорт (из корня проекта)
DATABASE_URL="postgres://postgres:postgres@localhost:5432/geosearch?sslmode=disable" \
  python3 scripts/ingest_osm.py
```

Конфигурация импорта: `scripts/ingest_config.json`

### Пересборка Docker

```bash
docker compose up -d --build
```

### Логи

```bash
# Все сервисы
docker compose logs -f

# Только API
docker compose logs -f api

# Только NLP
docker compose logs -f nlp
```

## Fallback парсер

При недоступности Python NLP сервиса Go-бэкенд использует встроенный regex-парсер (`internal/services/fallback_parser.go`). Он работает без ML, на основе словарей и регулярных выражений:

- **Категории:** кафе, ресторан, бар, фастфуд, парк
- **Фичи:** wifi, терраса, тишина, розетки, завтраки, парковка, романтика, семья, живая музыка
- **Намерения:** работа, завтрак, ужин, романтика
- **Расстояние:** рядом (500м), недалеко (1000м), далеко (5000м)
- **Метро:** regex для извлечения названия станции

## Добавление POI вручную

```sql
INSERT INTO pois (name, category, address, city, rating, review_count, price_level,
  has_wifi, has_terrace, is_quiet, geom)
VALUES (
  'Новое Кафе', 'cafe', 'ул. Пример 1', 'Москва', 4.5, 100, 2,
  true, true, true,
  ST_SetSRID(ST_MakePoint(37.60, 55.75), 4326)::geography
);
```

## Технологии

| Компонент | Технология | Версия |
|-----------|------------|--------|
| Язык бэкенда | Go | 1.25 |
| HTTP фреймворк | Gin | 1.12 |
| SQL драйвер | pgx/v5 | 5.5 |
| Кеш | go-redis | 9.21 |
| Язык NLP | Python | 3.12 |
| NLP фреймворк | FastAPI | 0.115 |
| БД | PostgreSQL | 16 |
| Гео-расширение | PostGIS | 3.4 |
| Кеширование | Redis | 7 |
| Контейнеры | Docker Compose | v2 |
