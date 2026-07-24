<div align="center">

# Geo Search

**Умный геопоиск мест на русском языке**

Пользователь пишет «тихое кафе с террасой рядом с метро» — система находит подходящие места с объяснениями.

[![Go](https://img.shields.io/badge/Go-1.23-00ADD8?logo=go)](https://go.dev)
[![Python](https://img.shields.io/badge/Python-3.12-3776AB?logo=python)](https://python.org)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-4169E1?logo=postgresql)](https://postgresql.org)
[![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?logo=docker)](https://docker.com)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

</div>

---

## Как это работает

```
Пользователь: "тихое кафе с террасой у метро Парк Культуры"
         │
         ▼
┌─────────────┐     ┌──────────────┐     ┌──────────────┐
│   Frontend   │────▶│  Go Backend  │────▶│  Python NLP  │
│  React/Vite  │     │    Gin       │     │   FastAPI    │
└─────────────┘     └──────┬───────┘     └──────────────┘
                           │                     │
                    ┌──────┴───────┐      Парсинг текста
                    │              │      в структурированные
              ┌─────▼────┐  ┌─────▼────┐   фильтры
              │PostgreSQL │  │  Redis    │
              │ + PostGIS │  │  Cache   │
              └──────────┘  └──────────┘
                    │
           Пространственный запрос
           + ранжирование по интенту
```

**NLP парсит запрос:**
```json
{
  "category": "cafe",
  "intent": "default",
  "features": { "terrace": true, "quiet": true },
  "location": { "metro": "Парк Культуры" },
  "radius": 500
}
```

**Go бэкенд строит SQL-запрос** с PostGIS `ST_DWithin()`, фильтрами по фичам, ранжированием по интенту и генерирует объяснение на русском.

---

## Быстрый старт

```bash
git clone https://github.com/qqwozz/geo-search-go.git
cd geo-search-go
docker compose up -d
```

| Сервис | URL | Описание |
|--------|-----|----------|
| Web UI | http://localhost:3000 | Карта + поиск |
| API | http://localhost:8080 | REST API |
| Swagger | http://localhost:8080/swagger/index.html | Документация API |
| NLP | http://localhost:8000 | Парсер текста |
| Postgres | localhost:5432 | БД |
| Redis | localhost:6379 | Кеш |

### Импорт данных

```bash
DATABASE_URL="postgres://postgres:postgres@localhost:5432/geosearch?sslmode=disable" \
  python3 scripts/ingest_osm.py
```

Данные загружаются из OpenStreetMap (кафе, рестораны, бары, фастфуд) в радиусе заданного bbox.

---

## API

### Поиск мест

```
GET /api/search?q=тихое+кафе+с+террасой&lat=55.755&lon=37.605
```

| Параметр | По умолчанию | Описание |
|----------|--------------|----------|
| `q` | — | Текстовый запрос (макс. 500 символов) |
| `lat`, `lon` | — | Центр поиска |
| `radius` | 2000 | Радиус в метрах |
| `limit` | 20 | Количество результатов (макс. 50) |
| `offset` | 0 | Смещение для пагинации |
| `sort` | relevance | `relevance` или `rating` |
| `city` | moscow | Город для NLP |

<details>
<summary>Пример ответа</summary>

```json
{
  "pois": [
    {
      "id": 3,
      "name": "Тихий Дворик",
      "category": "cafe",
      "address": "пер. Малый Казённый 7",
      "rating": 4.7,
      "review_count": 95,
      "features": {
        "wifi": true, "outlets": true, "terrace": false,
        "quiet": true, "romantic": false
      },
      "distance_meters": 414.34,
      "explanation": "Отличное место для работы: быстрый Wi-Fi, удобные розетки, очень тихо",
      "score": 10.12
    }
  ],
  "total": 4,
  "cached": false
}
```

</details>

### Автодополнение

```
GET /api/autocomplete?q=кафе
```

### Health check

```
GET /api/health
```

---

## Интенты и ранжирование

NLP определяет **интент** запроса и система ранжирует результаты по-разному:

| Интент | Пример запроса | Ключевые факторы |
|--------|----------------|------------------|
| **work** | «где поработать с ноутбуком» | Wi-Fi + розетки + тишина |
| **breakfast** | «завтрак рядом» | наличие завтрака + цена + рейтинг |
| **dinner** | «вечерний ужин» | рейтинг + цена + живая музыка |
| **romantic** | «ресторан для свидания» | романтика + тишина + рейтинг |
| **default** | «кафе с террасой» | рейтинг + расстояние + отзывы |

---

## Структура проекта

```
geo-search-go/
├── cmd/api/main.go              # Точка входа, Swagger, middleware
├── internal/
│   ├── config/                   # Конфигурация из env
│   ├── database/                 # pgx pool, Redis клиент
│   ├── handlers/                 # HTTP обработчики
│   │   ├── search.go             # GET /api/search
│   │   ├── autocomplete.go       # GET /api/autocomplete
│   │   └── health.go             # GET /api/health
│   ├── middleware/
│   │   └── ratelimit.go          # Token bucket rate limiter
│   ├── models/poi.go             # Структуры данных
│   └── services/
│       ├── search.go             # Оркестратор поиска
│       ├── nlp.go                # HTTP клиент к NLP
│       ├── cache.go              # Redis кеш (SHA-256 ключи, 5 мин TTL)
│       ├── fallback_parser.go    # Regex fallback парсер
│       ├── ranker.go             # Intent-based ранжирование
│       └── explainer.go          # Генерация объяснений на русском
├── nlp/
│   ├── main.py                   # FastAPI
│   ├── parser.py                 # Словарный парсинг
│   ├── dictionaries.py           # Русские словари
│   ├── location_parser.py        # Извлечение метро/улиц
│   └── tests/                    # Python тесты
├── frontend/
│   ├── src/
│   │   ├── components/           # React компоненты
│   │   │   ├── SearchBar         # Поиск с автодополнением
│   │   │   ├── MapView           # Leaflet карта
│   │   │   ├── POICard           # Карточка результата
│   │   │   ├── DetailModal       # Детали места
│   │   │   ├── ResultsList       # Список результатов
│   │   │   ├── QuickFilters      # Быстрые фильтры
│   │   │   └── Header            # Заголовок со статусом
│   │   └── storage.js            # LocalStorage (избранное, история)
│   └── Dockerfile                # nginx
├── scripts/
│   ├── ingest_osm.py             # Импорт из OpenStreetMap
│   └── ingest_config.json        # bbox и категории
├── migrations/
│   ├── 001_create_database.sql
│   ├── 002_create_pois_table.sql
│   └── 003_create_poi_indexes.sql
├── docs/                         # Swagger документация
├── Dockerfile                    # Multi-stage Go сборка
└── docker-compose.yml
```

---

## Технологии

| Компонент | Технология |
|-----------|------------|
| Бэкенд | Go 1.23, Gin, pgx/v5 |
| NLP | Python 3.12, FastAPI |
| БД | PostgreSQL 16 + PostGIS 3.4 |
| Кеш | Redis 7 |
| Frontend | React 18, Vite, Leaflet |
| Контейнеры | Docker Compose |
| Логирование | slog (structured JSON) |
| Документация | Swagger / OpenAPI |

---

## Конфигурация

| Переменная | По умолчанию | Описание |
|------------|--------------|----------|
| `PORT` | 8080 | Порт API |
| `DATABASE_URL` | postgres://postgres:postgres@localhost:5432/geosearch | PostgreSQL |
| `REDIS_URL` | redis://localhost:6379 | Redis |
| `NLP_SERVICE_URL` | http://localhost:8000 | Python NLP |
| `CORS_ORIGIN` | http://localhost:3000 | CORS origin |

---

## Разработка

```bash
# Go тесты
go test ./... -v

# Python тесты
cd nlp && pytest tests/ -v

# Пересборка Docker
docker compose up -d --build

# Логи
docker compose logs -f api
```

---

## Лицензия

MIT License — Copyright 2026 Dima Kiselev
