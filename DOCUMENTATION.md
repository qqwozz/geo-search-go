# Geo Search — Документация проекта

B2C умный геопоисковик для жителей города. Пользователь пишет запрос на естественном русском языке (например, «тихое кафе с террасой рядом с метро»), система находит подходящие места с объяснениями.

---

## Содержание

1. [Архитектура](#1-архитектура)
2. [Быстрый старт](#2-быстрый-старт)
3. [Структура проекта](#3-структура-проекта)
4. [Конфигурация](#4-конфигурация)
5. [API Reference](#5-api-reference)
6. [Модели данных](#6-модели-данных)
7. [Алгоритм работы поиска](#7-алгоритм-работы-поиска)
8. [NLP-сервис (Python)](#8-nlp-сервис-python)
9. [Fallback-парсер](#9-fallback-парсер)
10. [Ранжирование](#10-ранжирование)
11. [Генерация объяснений](#11-генерация-объяснений)
12. [Кеширование](#12-кеширование)
13. [Rate Limiting](#13-rate-limiting)
14. [База данных](#14-база-данных)
15. [Импорт данных из OSM](#15-импорт-данных-из-osm)
16. [Docker и деплой](#16-docker-и-деплой)
17. [Тестирование](#17-тестирование)
18. [Дорожная карта](#18-дорожная-карта)

---

## 1. Архитектура

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

### Компоненты

| Сервис | Порт | Стек | Описание |
|--------|------|------|----------|
| **api** | 8080 | Go 1.25, Gin, pgx/v5 | Основной бэкенд, REST API |
| **nlp** | 8000 | Python 3.12, FastAPI | Парсинг текста в фильтры |
| **postgres** | 5432 | PostgreSQL 16 + PostGIS 3.4 | Хранение POI, пространственные запросы |
| **redis** | 6379 | Redis 7 | Кеширование результатов |
| **frontend** | 3000 | nginx | Статический frontend |

### Поток запроса

1. Пользователь вводит текстовый запрос
2. Go-бэкенд валидирует параметры (q, lat, lon, radius, limit, sort)
3. Проверяется Redis-кеш — если результат есть, возвращается сразу
4. Запрос отправляется в Python NLP-сервис для парсинга
5. Если NLP недоступен — используется fallback regex-парсер
6. На основе NLP-фильтров строится SQL-запрос к PostGIS
7. Результаты ранжируются по intent (work/breakfast/romantic/default)
8. К каждому месту генерируется объяснение
9. Результат кешируется в Redis на 5 минут
10. Ответ возвращается клиенту

---

## 2. Быстрый старт

### Запуск через Docker

```bash
git clone <repo-url>
cd geo-search-go
docker compose up -d
```

После запуска доступны:
- API: `http://localhost:8080`
- NLP: `http://localhost:8000`
- Postgres: `localhost:5432`
- Redis: `localhost:6379`
- Frontend: `http://localhost:3000`

### Импорт данных

```bash
# Дождаться запуска postgres
docker compose up -d postgres

# Импорт POI из OpenStreetMap
docker compose exec api python scripts/ingest_osm.py
```

### Проверка работоспособности

```bash
curl http://localhost:8080/api/health
```

### Первый поиск

```bash
curl "http://localhost:8080/api/search?q=тихое+кафе+с+террасой&lat=55.755&lon=37.605&radius=2000"
```

### Локальная разработка (без Docker)

```bash
# PostgreSQL + PostGIS должны быть запущены на localhost:5432
# Redis — на localhost:6379

export DATABASE_URL="postgres://postgres:postgres@localhost:5432/geosearch?sslmode=disable"
export REDIS_URL="redis://localhost:6379"
export NLP_SERVICE_URL="http://localhost:8000"

# Go API
go run cmd/api/main.go

# Python NLP (отдельный терминал)
cd nlp
pip install -r requirements.txt
uvicorn main:app --host 0.0.0.0 --port 8000
```

---

## 3. Структура проекта

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
│   │   ├── search.go                # GET /api/search — хэндлер поиска
│   │   ├── autocomplete.go          # GET /api/autocomplete — подсказки
│   │   └── health.go                # GET /api/health — проверка здоровья
│   ├── middleware/
│   │   └── ratelimit.go             # Rate limiting по IP
│   ├── models/
│   │   └── poi.go                   # Структуры данных
│   └── services/
│       ├── search.go                # Оркестратор поиска (главный workflow)
│       ├── nlp.go                   # HTTP-клиент к Python NLP
│       ├── cache.go                 # Redis кеширование
│       ├── fallback_parser.go       # Regex-парсер (fallback)
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
│   ├── 003_create_poi_indexes.sql   # Основные индексы
│   └── 004_add_feature_indexes.sql  # Индексы по фичам
├── frontend/
│   ├── dist/                        # Собранный фронтенд
│   ├── Dockerfile
│   └── nginx.conf
├── docker-compose.yml
├── Dockerfile
├── .env.example
├── go.mod
└── go.sum
```

---

## 4. Конфигурация

### Env-переменные

| Переменная | По умолчанию | Описание |
|------------|--------------|----------|
| `PORT` | `8080` | Порт Go API |
| `DATABASE_URL` | `postgres://postgres:postgres@localhost:5432/geosearch?sslmode=disable` | URL подключения к PostgreSQL |
| `REDIS_URL` | `redis://localhost:6379` | URL подключения к Redis |
| `NLP_SERVICE_URL` | `http://localhost:8000` | URL Python NLP-сервиса |
| `CORS_ORIGIN` | `http://localhost:3000` | Разрешённый_ORIGIN для CORS |

### Создание .env файла

```bash
cp .env.example .env
# Отредактируйте под свою среду
```

---

## 5. API Reference

### GET /api/search

Поиск мест по естественному запросу.

**Query Parameters:**

| Параметр | Тип | Обязательный | По умолчанию | Описание |
|----------|-----|--------------|--------------|----------|
| `q` | string | Да | — | Текстовый запрос на русском языке (макс. 500 символов) |
| `lat` | float | Да | — | Широта центра поиска |
| `lon` | float | Да | — | Долгота центра поиска |
| `radius` | int | Нет | `2000` | Радиус поиска в метрах |
| `limit` | int | Нет | `20` | Количество результатов (макс. 50) |
| `sort` | string | Нет | `relevance` | Сортировка: `relevance`, `distance`, `rating` |

**Пример запроса:**

```bash
curl "http://localhost:8080/api/search?q=уютное+кафе+где+можно+поработать&lat=55.755&lon=37.605&radius=2000"
```

**Ответ (200 OK):**

```json
{
  "pois": [
    {
      "id": 3,
      "name": "Тихий Дворик",
      "name_en": null,
      "category": "cafe",
      "subcategory": null,
      "address": "пер. Малый Казённый 7",
      "city": "Москва",
      "phone": null,
      "website": null,
      "opening_hours": null,
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
      "score": 10.12,
      "last_updated": "2025-01-15T10:30:00Z"
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
| `pois[].explanation` | Человеческое объяснение на русском языке |
| `pois[].distance_meters` | Расстояние от центра поиска в метрах |
| `pois[].features` | Доступные удобства места |
| `total` | Количество результатов |
| `cached` | Взят ли результат из кеша |

**Ошибки:**

| Код | Причина |
|-----|---------|
| 400 | Параметр `q` пустой, `lat`/`lon` отсутствуют или вне диапазона, `q` длиннее 500 символов |
| 429 | Превышен лимит запросов (rate limit) |
| 500 | Внутренняя ошибка сервера |

---

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

**Ответ (200 OK):**

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

Если параметр `q` пустой — возвращаются 8 самых популярных запросов.

---

### GET /api/health

Проверка работоспособности всех сервисов.

**Ответ (200 OK):**

```json
{
  "status": "ok",
  "postgres": "ok",
  "redis": "ok",
  "nlp": "ok"
}
```

**Ответ (503 Service Unavailable) — если NLP недоступен:**

```json
{
  "status": "degraded",
  "postgres": "ok",
  "redis": "ok",
  "nlp": "error"
}
```

Если PostgreSQL или Redis недоступны — возвращается 500.

---

### POST /parse (NLP-сервис, порт 8000)

Парсинг текстового запроса в структурированные фильтры.

**Request Body:**

```json
{
  "text": "тихое кафе с террасой рядом с метро Парк Культуры",
  "city": "moscow"
}
```

**Ответ (200 OK):**

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

**Поля NLP-ответа:**

| Поле | Тип | Описание |
|------|-----|----------|
| `category` | string | Категория места: `cafe`, `restaurant`, `bar`, `fast_food`, `park` |
| `intent` | string | Намерение: `work`, `breakfast`, `dinner`, `romantic`, `default` |
| `features` | object | Обнаруженные фичи (wifi, terrace, quiet, outlets, breakfast, parking, romantic, family, live_music) |
| `location` | object | Подсказка о локации: `metro`, `street`, `area` |
| `radius` | int | Радиус поиска в метрах (извлечённый из текста) |
| `radius_raw` | string | Исходное слово для радиуса ("рядом", "недалеко", etc.) |
| `sort_by` | string | Рекомендуемая сортировка |

---

## 6. Модели данных

### POI (Point of Interest)

```go
type POI struct {
    ID             int       `json:"id"`
    Name           string    `json:"name"`
    NameEN         *string   `json:"name_en,omitempty"`
    Category       string    `json:"category"`
    Subcategory    *string   `json:"subcategory,omitempty"`
    Address        *string   `json:"address,omitempty"`
    City           *string   `json:"city,omitempty"`
    Phone          *string   `json:"phone,omitempty"`
    Website        *string   `json:"website,omitempty"`
    OpeningHours   *string   `json:"opening_hours,omitempty"`
    Rating         float64   `json:"rating"`
    ReviewCount    int       `json:"review_count"`
    PriceLevel     int       `json:"price_level"`
    Features       Features  `json:"features"`
    NoiseLevel     string    `json:"noise_level"`
    Lat            float64   `json:"lat"`
    Lon            float64   `json:"lon"`
    DistanceMeters float64   `json:"distance_meters,omitempty"`
    Explanation    string    `json:"explanation,omitempty"`
    Score          float64   `json:"score,omitempty"`
    LastUpdated    time.Time `json:"last_updated"`
}
```

### Features (Удобства)

```go
type Features struct {
    Wifi           bool `json:"wifi"`
    Outlets        bool `json:"outlets"`
    Terrace        bool `json:"terrace"`
    Parking        bool `json:"parking"`
    LiveMusic      bool `json:"live_music"`
    Breakfast      bool `json:"breakfast"`
    Quiet          bool `json:"quiet"`
    FamilyFriendly bool `json:"family_friendly"`
    Romantic       bool `json:"romantic"`
    DogFriendly    bool `json:"dog_friendly"`
}
```

### SearchRequest

```go
type SearchRequest struct {
    Query  string  `json:"query" form:"q"`
    Lat    float64 `json:"lat" form:"lat"`
    Lon    float64 `json:"lon" form:"lon"`
    Radius int     `json:"radius" form:"radius"`
    Limit  int     `json:"limit" form:"limit"`
    Sort   string  `json:"sort" form:"sort"`
}
```

### SearchResponse

```go
type SearchResponse struct {
    POIs   []POI `json:"pois"`
    Total  int   `json:"total"`
    Center struct {
        Lat float64 `json:"lat"`
        Lon float64 `json:"lon"`
    } `json:"center"`
    Query  string `json:"query"`
    Cached bool   `json:"cached"`
}
```

### NLPResponse

```go
type NLPResponse struct {
    Category  string          `json:"category"`
    Intent    string          `json:"intent"`
    Features  map[string]bool `json:"features"`
    Location  *LocationHint   `json:"location"`
    Radius    int             `json:"radius"`
    RadiusRaw string          `json:"radius_raw"`
    SortBy    string          `json:"sort_by"`
}
```

---

## 7. Алгоритм работы поиска

Основная логика находится в `internal/services/search.go`.

### Пошаговый流程

```
1. Валидация параметров
       │
2. Проверка Redis-кеша ────命中────▶ Возврат кешированного ответа
       │ (промах)
3. Отправка запроса в NLP-сервис
       │
       ├── NLP доступен ──────────▶ Получение NLPResponse
       │
       └── NLP недоступен ────────▶ FallbackParse() (regex-парсер)
       │
4. Построение SQL-запроса к PostGIS
   ├── ST_DWithin() для пространственной фильтрации
   ├── WHERE category = ...
   ├── WHERE has_wifi = TRUE (если запрошено)
   └── ORDER BY расстояние / рейтинг
       │
5. Ранжирование по intent (RankByIntent)
       │
6. Генерация объяснений (GenerateExplanation)
       │
7. Сортировка по score
       │
8. Кеширование в Redis (TTL 5 минут)
       │
9. Возврат ответа
```

### Построение SQL-запроса

Запрос динамически строится в `queryPOIs()`:

```sql
SELECT
    id, name, name_en, category, subcategory, address, city, phone, website,
    opening_hours::text, rating, review_count, price_level,
    has_wifi, has_outlets, has_terrace, has_parking, has_live_music, has_breakfast,
    is_quiet, is_family_friendly, is_romantic, is_dog_friendly,
    noise_level,
    ST_Y(geom::geometry) as lat,
    ST_X(geom::geometry) as lon,
    ST_Distance(geom, ST_MakePoint($1, $2)::geography) as distance
FROM pois
WHERE ST_DWithin(geom, ST_MakePoint($1, $2)::geography, $3)
  AND category = $4
  AND has_wifi = TRUE
  -- ... другие фильтры
ORDER BY geom <-> ST_MakePoint($1, $2)::geography
LIMIT $N
```

### Защита от SQL-инъекций

Все колонки проверяются через whitelist `allowedColumns`:

```go
var allowedColumns = map[string]bool{
    "category":           true,
    "has_wifi":           true,
    "has_outlets":        true,
    // ... 14 колонок всего
}
```

Параметры передаются через параметризованные запросы `$1`, `$2` — конкатенация строк не используется.

---

## 8. NLP-сервис (Python)

FastAPI-приложение в директории `nlp/`.

### Файлы

| Файл | Описание |
|------|----------|
| `main.py` | FastAPI-приложение, эндпоинты `/parse` и `/health` |
| `parser.py` | Основная логика парсинга русского текста |
| `dictionaries.py` | Словари категорий, фич, намерений |
| `location_parser.py` | Извлечение информации о локации (метро, улицы) |

### Запуск

```bash
cd nlp
pip install -r requirements.txt
uvicorn main:app --host 0.0.0.0 --port 8000
```

### Тесты

```bash
cd nlp
pytest tests/ -v
```

---

## 9. Fallback-парсер

При недоступности Python NLP-сервиса Go-бэкенд использует встроенный regex-парсер (`internal/services/fallback_parser.go`).

### Возможности

| Аспект | Поддержка |
|--------|-----------|
| **Категории** | кафе, ресторан, бар, фастфуд, парк |
| **Фичи** | wifi, терраса, тишина, розетки, завтраки, парковка, романтика, семья, живая музыка |
| **Намерения** | работа, завтрак, ужин, романтика |
| **Расстояние** | рядом (500м), недалеко (1000м), далеко (5000м) |
| **Метро** | Regex для извлечения названия станции |

### Алгоритм

1. Текст приводится к нижнему регистру
2. Поиск категории по ключевым словам из `categoryKeywords`
3. Поиск фич по ключевым словам из `featureKeywords`
4. Определение намерения по ключевым словам из `intentKeywords`
5. Определение радиуса по словам-индикаторам расстояния
6. Извлечение названия метро через regex

### Примеры

| Запрос | Категория | Фичи | Намерение | Радиус |
|--------|-----------|-------|-----------|--------|
| «тихое кафе с террасой» | cafe | quiet, terrace | default | 2000 |
| «бар рядом» | bar | — | default | 500 |
| «кафе где можно поработать с ноутбуком» | cafe | — | work | 2000 |
| «ресторан для свидания» | restaurant | romantic | romantic | 2000 |

---

## 10. Ranжирование

Файл: `internal/services/ranker.go`

Алгоритм зависит от намерения (intent), определённого NLP:

### Work (работа)

```
score = (wifi × 3) + (outlets × 3) + (quiet × 2) + (rating/5 × 2) + (1/distance × 100) - (live_music × 3)
```

- Wi-Fi и розетки — критичны для работы (+3 каждый)
- Тишина — важна (+2)
- Живая музыка — мешает (-3)
- Близость к пользователю — сильно влияет

### Breakfast (завтрак)

```
score = (breakfast × 5) + (rating/5 × 3) + (price_level ≤ 2 × 2) + (1/distance × 50)
```

- Наличие завтраков — главный фактор (+5)
- Доступные цены — важны (+2)

### Romantic (романтика)

```
score = (romantic × 4) + (rating/5 × 2) + (quiet × 2) + (1/distance × 30)
```

- Романтическая атмосфера — главный фактор (+4)
- Тишина — важна (+2)

### Default (общий)

```
score = (rating/5 × 5) + (1/distance × 50) + (review_count/100 × 2)
```

- Рейтинг — главный фактор (+5)
- Популярность (отзывы) — важна (+2)

### Расстояние

Для всех intent расстояние влияет через формулу `1/distance`, но с разными коэффициентами:
- Work: ×100
- Breakfast: ×50
- Romantic: ×30
- Default: ×50

Расстояние ограничено снизу 10м (защита от деления на 0).

---

## 11. Генерация объяснений

Файл: `internal/services/explainer.go`

К каждому результату добавляется текстовое объяснение на русском языке.

### Шаблоны по intent

| Intent | Шаблон | Пример |
|--------|--------|--------|
| work | «Отличное место для работы: {features}» | «Отличное место для работы: быстрый Wi-Fi, удобные розетки» |
| breakfast | «Завтраки есть в меню» + «доступные цены» | |
| romantic | «Идеально для свидания: {features}» | «Идеально для свидания: романтичная атмосфера, тихо» |
| default | «Высокий рейтинг X.X» + «много отзывов (N)» | |

### Дополнительные фичи

В конце объяснения добавляются «bonus»-фичи:
- Терраса, парковка, семейное — если есть

### Fallback

Если объяснение пустое — «Хороший вариант поблизости».

---

## 12. Кеширование

Файл: `internal/services/cache.go`

### Механизм

1. **Генерация ключа:** SHA-256 хеш от строки `query:lat:lon:radius:limit:sort`
2. **Хранение:** JSON-сериализованный `SearchResponse` в Redis
3. **TTL:** 5 минут
4. **Префикс ключа:** `search:`

### Проверка кеша

При каждом поисковом запросе:
1. Генерируется ключ на основе параметров запроса
2. Если результат найден в Redis — возвращается с `cached: true`
3. Если нет — выполняется полный цикл поиска и результат кешируется

### Ограничения

- Кеш хранит только полные ответы (включая POI с score и explanation)
- При изменении данных в БД кеш автоматически устареется через 5 минут

---

## 13. Rate Limiting

Файл: `internal/middleware/ratelimit.go`

### Настройки

- **Скорость:** 0.5 запросов в секунду (30 запросов в минуту)
- **Burst:** 10 запросов одновременно
- **Очистка:** неактивные IP удаляются через 3 минуты

### Реализация

- Token bucket алгоритм через `golang.org/x/time/rate`
- Каждый IP получает свой лимитер
- При превышении лимита возвращается 429 Too Many Requests

### Ответ при превышении

```json
{
  "error": "rate limit exceeded",
  "retry_after": "1s"
}
```

---

## 14. База данных

### PostgreSQL + PostGIS

Версия: PostgreSQL 16 с расширением PostGIS 3.4.

### Таблица `pois`

```sql
CREATE TABLE pois (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    name_en VARCHAR(255),
    category VARCHAR(50) NOT NULL,
    subcategory VARCHAR(50),
    address TEXT,
    city VARCHAR(100),
    phone VARCHAR(20),
    website VARCHAR(500),
    opening_hours JSONB,
    rating DECIMAL(2,1) DEFAULT 0,
    review_count INTEGER DEFAULT 0,
    price_level INTEGER DEFAULT 2,
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

### Индексы

```sql
-- Пространственный индекс (основной для гео-запросов)
CREATE INDEX idx_pois_geom ON pois USING GIST (geom);

-- Индексы для фильтрации
CREATE INDEX idx_pois_category ON pois (category);
CREATE INDEX idx_pois_rating ON pois (rating DESC);
CREATE INDEX idx_pois_city ON pois (city);

-- Частичные индексы для фич (только для TRUE значений)
CREATE INDEX idx_pois_wifi ON pois (has_wifi) WHERE has_wifi = TRUE;
CREATE INDEX idx_pois_terrace ON pois (has_terrace) WHERE has_terrace = TRUE;
CREATE INDEX idx_pois_quiet ON pois (is_quiet) WHERE is_quiet = TRUE;
CREATE INDEX idx_pois_breakfast ON pois (has_breakfast) WHERE has_breakfast = TRUE;
```

### Добавление POI вручную

```sql
INSERT INTO pois (name, category, address, city, rating, review_count, price_level,
  has_wifi, has_terrace, is_quiet, geom)
VALUES (
  'Новое Кафе', 'cafe', 'ул. Пример 1', 'Москва', 4.5, 100, 2,
  true, true, true,
  ST_SetSRID(ST_MakePoint(37.60, 55.75), 4326)::geography
);
```

---

## 15. Импорт данных из OSM

### Скрипт

Файл: `scripts/ingest_osm.py`

### Зависимости

```bash
pip install overpy psycopg2-binary
```

### Запуск

```bash
DATABASE_URL="postgres://postgres:postgres@localhost:5432/geosearch?sslmode=disable" \
  python3 scripts/ingest_osm.py
```

### Конфигурация

Файл: `scripts/ingest_config.json`

### Источники данных

- **OpenStreetMap** — основной источник POI (кафе, рестораны, бары)
- **Ручное обогащение** — администратор добавляет фото, часы работы, контакты

### Периодичность обновления

- OSM-дамп — раз в неделю (автоматически)
- Ручные правки — моментально

---

## 16. Docker и деплой

### docker-compose.yml

Определяет 5 сервисов:

| Сервис | Образ | Ресурсы |
|--------|-------|---------|
| postgres | postgis/postgis:16-3.4 | 512M RAM, 1 CPU |
| redis | redis:7-alpine | 128M RAM, 0.5 CPU |
| api | Сборка из Dockerfile | 256M RAM, 1 CPU |
| nlp | Сборка из nlp/Dockerfile | 256M RAM, 0.5 CPU |
| frontend | Сборка из frontend/Dockerfile | — |

### Dockerfile (Go API)

Многостадийная сборка:
1. **builder:** `golang:1.25-alpine` — компиляция Go-бинарника
2. **runtime:** `alpine:3.19` — минимальный образ с бинарником

### Health checks

Все сервисы имеют health checks:
- postgres: `pg_isready -U postgres`
- redis: `redis-cli ping`
- api: `wget --spider -q http://localhost:8080/api/health`
- nlp: Python-скрипт для проверки `/health`

### Полезные команды

```bash
# Запуск всех сервисов
docker compose up -d

# Пересборка
docker compose up -d --build

# Логи всех сервисов
docker compose logs -f

# Логи только API
docker compose logs -f api

# Остановка и удаление данных
docker compose down -v

# Проверка здоровья
docker compose ps
```

---

## 17. Тестирование

### Go тесты

```bash
go test ./internal/services/... -v
```

Тесты покрывают:
- `fallback_parser_test.go` — тесты regex-парсера

### Python тесты

```bash
cd nlp && pytest tests/ -v
```

Тесты покрывают:
- `test_parser.py` — тесты NLP-парсера
- `test_location.py` — тесты парсера локации

### Ручное тестирование

```bash
# Health check
curl http://localhost:8080/api/health

# Поиск кафе
curl "http://localhost:8080/api/search?q=тихое+кафе+с+террасой&lat=55.755&lon=37.605&radius=2000"

# Автокомплит
curl "http://localhost:8080/api/autocomplete?q=кафе"

# NLP парсинг напрямую
curl -X POST http://localhost:8000/parse \
  -H "Content-Type: application/json" \
  -d '{"text":"тихое кафе с террасой"}'
```

---

## 18. Дорожная карта

### Phase 1: Core Backend (недели 1-3) — ВЫПОЛНЕНО

- [x] Go бэкенд: /search, /autocomplete, /health
- [x] Python NLP: парсинг русского текста
- [x] Fallback regex-парсер
- [x] Intent-based ранжирование
- [x] Генерация объяснений
- [x] Redis кеширование (TTL 5 мин)
- [x] Rate limiting (30 req/min)
- [x] OSM ингест
- [x] Docker Compose
- [x] Тесты Go + Python

### Phase 2: Web Frontend (неделя 4)

- [ ] React + Vite
- [ ] Поисковая строка с автокомплитом
- [ ] Чипы быстрых фильтров
- [ ] Карта Leaflet/OpenStreetMap
- [ ] Карточки результатов
- [ ] Модалка детальной информации
- [ ] Кнопка "Поделиться"
- [ ] Избранное + история (LocalStorage)
- [ ] Геолокация через браузерный API

### Phase 3: Telegram Bot (неделя 5)

- [ ] python-telegram-bot
- [ ] Текстовый запрос → топ-5 результатов
- [ ] "Показать на карте" / "Маршрут"
- [ ] Inline query поддержка

### Phase 4: Обогащение (месяцы 2-3)

- [ ] Голосовой поиск (Web Speech API)
- [ ] Персонализация
- [ ] Мультиязычность
- [ ] Обратная связь от пользователей
- [ ] Автоматическое обогащение тегами

### Phase 5: Масштабирование (месяцы 4-6)

- [ ] Мобильное приложение (React Native / Flutter)
- [ ] Push-уведомления
- [ ] Партнерские интеграции
- [ ] Рекламные карточки (CPM)
- [ ] Премиум-функции

---

## Метрики успеха

| Метрика | Целевое значение | Срок |
|---------|-----------------|------|
| Время ответа API | < 500 мс (p95) | Phase 1 |
| Точность NLP | > 85% | Phase 1 |
| Пользователи в неделю | 1000+ | Phase 3 |
| Повторные визиты | > 30% | Phase 4 |
| CTR на результаты | > 40% | Phase 2 |
| Доля запросов с геолокацией | > 60% | Phase 2 |

---

## Технологии

| Компонент | Технология | Версия |
|-----------|------------|--------|
| Язык бэкенда | Go | 1.25 |
| HTTP фреймворк | Gin | 1.12 |
| SQL драйвер | pgx/v5 | 5.5 |
| Кеш клиент | go-redis | 9.21 |
| Rate limiting | golang.org/x/time/rate | 0.11 |
| Язык NLP | Python | 3.12 |
| NLP фреймворк | FastAPI | 0.115 |
| БД | PostgreSQL | 16 |
| Гео-расширение | PostGIS | 3.4 |
| Кеширование | Redis | 7 |
| Контейнеры | Docker Compose | v2 |
