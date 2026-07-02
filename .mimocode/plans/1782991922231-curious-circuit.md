# Phase 1: Core Backend — Детальный план

## Задача
Полностью заменить B2B-прототип (недвижимость) на B2C-ядро (POI-поиск). К концу Phase 1 работает: Go API + Python NLP + PostGIS + Redis + Docker Compose + OSM-данные.

---

## Шаг 1: Миграции (0.5 дня)

### Удалить
- `migrations/002_create_properties_table.sql`
- `migrations/003_add_indexes.sql`
- `migrations/seed.sql`

### Создать `migrations/002_create_pois_table.sql`
Таблица `pois`:
- id SERIAL PK
- name VARCHAR(255), name_en VARCHAR(255)
- category VARCHAR(50) — cafe, restaurant, bar, park, gym
- subcategory VARCHAR(50) — coffee_shop, fine_dining, fast_food
- address TEXT, city VARCHAR(100), phone VARCHAR(20), website VARCHAR(500)
- opening_hours JSONB — {"mon":"08:00-22:00", ...}
- rating DECIMAL(2,1), review_count INT, price_level INT (1-3)
- Feature flags: has_wifi, has_outlets, has_terrace, has_parking, has_live_music, has_breakfast, is_quiet, is_family_friendly, is_romantic, is_dog_friendly
- noise_level VARCHAR(10) — low/medium/high
- osm_id BIGINT UNIQUE, osm_type VARCHAR(10)
- source VARCHAR(50) DEFAULT 'osm'
- geom GEOGRAPHY(POINT, 4326) NOT NULL
- created_at, last_updated

### Создать `migrations/003_create_poi_indexes.sql`
- GiST на geom
- B-tree на category, rating DESC, city
- Partial indexes на wifi, terrace, quiet, breakfast

---

## Шаг 2: Go-бэкенд — конфиг и модели (1 день)

### `internal/config/config.go`
Структура Config с полями: Port, DatabaseURL, RedisURL, NLPServiceURL, CORSOrigin.
Функция `Load()` читает из env-переменных, дефолты если пусто.

### `internal/models/poi.go`
Структуры:
- `POI` — все поля таблицы + DistanceMeters, Explanation, Score
- `Features` — булевы флаги amenities
- `SearchRequest` — Query, Lat, Lon, Radius, Limit, Sort
- `SearchResponse` — POIs []POI, Total, Center, Query, Cached bool
- `NLPResponse` — Category, Intent, Features map, Location, Radius, SortBy
- `LocationHint` — Metro, Street, Area, Lat, Lon

---

## Шаг 3: Go-бэкенд — БД и Redis (0.5 дня)

### `internal/database/postgres.go`
- `InitPool(databaseURL string) (*pgxpool.Pool, error)`
- Проверка соединения через Ping

### `internal/database/redis.go`
- `InitClient(redisURL string) (*redis.Client, error)`
- Проверка соединения

---

## Шаг 4: Go-бэкенд — сервисы (4 дня)

### `internal/services/nlp.go`
- `ParseQuery(pool *pgxpool.Pool, text, city string) (*NLPResponse, error)`
- HTTP POST к Python NLP `/parse` с таймаутом 2 секунды
- При ошибке/таймауте → вызов fallback парсера

### `internal/services/fallback_parser.go`
Словари на русском:
- categoryKeywords: кафе→cafe, ресторан→restaurant, бар→bar, парк→park
- featureKeywords: вайфай→wifi, терраса→terrace, тихо→quiet, розетки→outlets, завтрак→breakfast
- intentKeywords: работ→work, завтрак→breakfast, ужин→dinner, романтич→romantic
- distanceKeywords: рядом→500, недалеко→1000, близко→500, далеко→5000
- metroPattern: regex `метро\s+([А-Яа-яёЁ\s]+?)(?:\s*[,.]|\s*$)`
- Функция `FallbackParse(text string) *NLPResponse`

### `internal/services/search.go`
Основной оркестратор:
1. Проверить Redis кеш (ключ `search:{query_hash}:{lat}:{lon}:{radius}`)
2. Если кеш — вернуть
3. Вызвать NLP (или fallback)
4. Разрешить location hints (метро → координаты через PostGIS)
5. Построить динамический SQL: WHERE ST_DWithin + category + feature flags
6. Выполнить запрос
7. Ранжировать через ranker.go
8. Сгенерировать объяснения
9. Записать в Redis (TTL 300 сек)

### `internal/services/ranker.go`
4 формулы:
- **work**: wifi×3 + outlets×3 + quiet×2 + (rating/5)×2 + (1/dist)×100 - music×3
- **breakfast**: breakfast×5 + (rating/5)×3 + (price<2)×2 + (1/dist)×50
- **romantic**: romantic×4 + rating×2 + quiet×2 + (1/dist)×30
- **default**: (rating/5)×5 + (1/dist)×50 + (review_count/100)×2

### `internal/services/explainer.go`
Генерация строк на русском:
- "Отличное место для работы: быстрый Wi-Fi, удобные розетки и очень тихо"
- "Красивая терраса с видом, но может быть шумно в выходные"
- Логика: на основе тегов intent + features

### `internal/services/cache.go`
- `GetCache(client *redis.Client, key string) ([]byte, error)`
- `SetCache(client *redis.Client, key string, data []byte, ttl time.Duration) error`
- Ключ = hex(SHA256(query+lat+lon+radius))

---

## Шаг 5: Go-бэкенд — хэндлеры (1 день)

### `internal/handlers/search.go`
`GET /api/search?q=...&lat=...&lon=...&radius=...&limit=...&sort=...`
- Валидация: q непустой, < 500 символов; lat/lon обязательны
- Rate limiting: 30 req/min per IP (через Redis или in-memory)
- Вызов search service
- Возврат JSON

### `internal/handlers/autocomplete.go`
`GET /api/autocomplete?q=...`
- Поиск по популярным запросам/категориям
- Возврат до 8 подсказок

### `internal/handlers/health.go`
`GET /api/health`
- Проверка Postgres ping + Redis ping

---

## Шаг 6: Переписать main.go (0.5 дня)

`cmd/api/main.go`:
1. Загрузить конфиг через config.Load()
2. Инициализировать pgx pool + Redis
3. Зарегистрировать роуты
4. CORS middleware
5. Serve static из frontend/dist/
6. Graceful shutdown (signal.Notify SIGINT/SIGTERM)

---

## Шаг 7: Python NLP модуль (2 дня)

### Структура `nlp/`
```
nlp/
├── main.py              — FastAPI, POST /parse, GET /health
├── parser.py            — parse_query(text, city) → NLPResponse JSON
├── dictionaries.py      — CATEGORIES, FEATURES, INTENTS, DISTANCE словари
├── location_parser.py   — extract_metro(text), extract_street(text)
├── requirements.txt     — fastapi, uvicorn, pydantic
├── Dockerfile
└── tests/
    ├── test_parser.py
    └── test_location.py
```

### `nlp/parser.py` алгоритм
1. Lowercase + tokenize
2. Match categories (longest match first)
3. Match features
4. Match intents
5. Extract metro (regex)
6. Extract distance hints
7. Return JSON: category, intent, features, location, radius, sort_by

---

## Шаг 8: OSM ингест (1 день)

### `scripts/ingest_osm.py`
- Библиотека `overpy` для Overpass API
- Запрос bbox для Москвы (или настраиваемый)
- Маппинг OSM тегов → наша схема:
  - amenity=cafe → category=cafe
  - outdoor_seating=yes → has_terrace=true
  - internet_access=wlan → has_wifi=true
  - smoking=no → is_quiet=true (эвристика)
- Bulk insert через psycopg2

### `scripts/ingest_config.json`
```json
{
  "city": "moscow",
  "bbox": [37.0, 55.5, 38.0, 56.0],
  "categories": ["cafe", "restaurant", "bar", "fast_food"],
  "batch_size": 500
}
```

---

## Шаг 9: Docker Compose (0.5 дня)

### `docker-compose.yml`
5 сервисов:
- **postgres**: postgis/postgis:16-3.4, порт 5432, volume pgdata, init-скрипты из migrations/
- **redis**: redis:7-alpine, порт 6379
- **api**: Go бэкенд, порт 8080, depends_on postgres+redis, env_file .env
- **nlp**: Python NLP, порт 8000
- **frontend**: (заглушка на Phase 2), порт 3000

### `Dockerfile` (Go)
Multi-stage: golang:1.23-alpine → alpine:3.19

### `nlp/Dockerfile`
python:3.12-slim → pip install → uvicorn

---

## Шаг 10: Тесты (2 дня)

### Go тесты
- `internal/services/fallback_parser_test.go` — 15-20 кейсов парсинга русского текста
- `internal/services/ranker_test.go` — проверка 4 формул ранжирования

### Python тесты
- `nlp/tests/test_parser.py` — 10-15 кейсов NLP парсинга
- `nlp/tests/test_location.py` — извлечение метро/улицы

---

## Шаг 11: Верификация

```bash
# Запуск
docker compose up -d

# Ингест OSM данных
docker compose exec api python scripts/ingest_osm.py

# Тест NLP напрямую
curl -X POST http://localhost:8000/parse \
  -H "Content-Type: application/json" \
  -d '{"text":"тихое кафе с террасой рядом с метро Парк Культуры"}'

# Тест поиска
curl "http://localhost:8080/api/search?q=тихое+кафе+с+террасой&lat=55.75&lon=37.62&radius=1000"

# Тест кеша (второй запрос быстрее)
time curl "http://localhost:8080/api/search?q=тихое+кафе+с+террасой&lat=55.75&lon=37.62&radius=1000"

# Go тесты
go test ./internal/services/...

# Python тесты
cd nlp && pytest tests/
```

---

## Файловый манифест Phase 1

| Действие | Путь | Описание |
|----------|------|----------|
| DELETE | `migrations/002_create_properties_table.sql` | Старая схема |
| DELETE | `migrations/003_add_indexes.sql` | Старые индексы |
| DELETE | `migrations/seed.sql` | Старые данные |
| CREATE | `migrations/002_create_pois_table.sql` | POI таблица |
| CREATE | `migrations/003_create_poi_indexes.sql` | POI индексы |
| REWRITE | `cmd/api/main.go` | Новый bootstrap |
| CREATE | `internal/config/config.go` | Конфиг из env |
| CREATE | `internal/models/poi.go` | POI типы |
| CREATE | `internal/database/postgres.go` | pgx pool |
| CREATE | `internal/database/redis.go` | Redis клиент |
| CREATE | `internal/handlers/search.go` | /search |
| CREATE | `internal/handlers/autocomplete.go` | /autocomplete |
| CREATE | `internal/handlers/health.go` | /health |
| CREATE | `internal/services/search.go` | Оркестратор |
| CREATE | `internal/services/nlp.go` | NLP клиент |
| CREATE | `internal/services/cache.go` | Redis кеш |
| CREATE | `internal/services/fallback_parser.go` | Regex парсер |
| CREATE | `internal/services/ranker.go` | Ранжирование |
| CREATE | `internal/services/explainer.go` | Объяснения |
| CREATE | `nlp/main.py` | FastAPI |
| CREATE | `nlp/parser.py` | NLP логика |
| CREATE | `nlp/dictionaries.py` | Словари |
| CREATE | `nlp/location_parser.py` | Парсер локации |
| CREATE | `nlp/requirements.txt` | Python deps |
| CREATE | `nlp/Dockerfile` | Python контейнер |
| CREATE | `nlp/tests/test_parser.py` | NLP тесты |
| CREATE | `nlp/tests/test_location.py` | Тесты локации |
| CREATE | `scripts/ingest_osm.py` | OSM ингест |
| CREATE | `scripts/ingest_config.json` | Конфиг ингеста |
| CREATE | `docker-compose.yml` | Оркестрация |
| CREATE | `Dockerfile` | Go контейнер |
| CREATE | `.env.example` | Шаблон конфига |
| CREATE | `internal/services/fallback_parser_test.go` | Go тесты |
| CREATE | `internal/services/ranker_test.go` | Тесты ранжирования |
| CREATE | `roadmap.md` | ROADMAP проекта |

## Оценка трудозатрат

| Задача | Дни |
|--------|-----|
| Миграции | 0.5 |
| Go конфиг + модели | 1 |
| Go БД + Redis | 0.5 |
| Go сервисы (search, nlp, cache, ranker, explainer, fallback) | 4 |
| Go хэндлеры | 1 |
| main.go rewrite | 0.5 |
| Python NLP модуль | 2 |
| OSM ингест | 1 |
| Docker Compose + Dockerfiles | 0.5 |
| Тесты | 2 |
| Интеграция и отладка | 1.5 |
| **Итого** | **14.5 дней (~3 недели)** |

---
---

# ROADMAP проекта

## Визия
B2C умный геопоисковик для жителей города. Пользователь пишет на естественном языке, получает персонализированные рекомендации мест с объяснениями.

## Фазы

### Phase 1: Core Backend (недели 1-3)
**Цель:** Рабочий Go API + Python NLP + БД + Docker
- [ ] Замена схемы БД: properties → pois
- [ ] Go бэкенд: /search, /autocomplete, /health
- [ ] Python NLP: парсинг русского текста → JSON фильтры
- [ ] Fallback regex-парсер в Go
- [ ] Intent-based ранжирование (work, breakfast, romantic, default)
- [ ] Генерация объяснений к результатам
- [ ] Redis кеширование (TTL 5 мин)
- [ ] Rate limiting (30 req/min)
- [ ] OSM ингест (cafes, restaurants, bars)
- [ ] Docker Compose (postgres+postgis, redis, api, nlp)
- [ ] Тесты Go + Python

### Phase 2: Web Frontend (неделя 4)
**Цель:** Веб-интерфейс с картой и поиском
- [ ] React + Vite
- [ ] Поисковая строка с автокомплитом
- [ ] Чипы быстрых фильтров (☕ Кофе, 🍳 Завтрак, 💼 Работа, 🌿 Терраса)
- [ ] Карта Leaflet/OpenStreetMap с маркерами
- [ ] Карточки результатов (рейтинг, расстояние, теги, объяснение)
- [ ] Модалка детальной информации (адрес, телефон, часы, маршрут)
- [ ] Кнопка "Поделиться" (URL с query params)
- [ ] Избранное + история (LocalStorage)
- [ ] Геолокация через браузерный API
- [ ] Nginx в Docker для production

### Phase 3: Telegram Bot (неделя 5)
**Цель:** Telegram-бот для мобильных пользователей
- [ ] python-telegram-bot
- [ ] /start, /help команды
- [ ] Текстовый запрос → топ-5 результатов с кнопками
- [ ] "Показать на карте" → ссылка на Яндекс.Карты
- [ ] "Маршрут" → навигация
- [ ] "Сохранить" → избранное в сессии
- [ ] "Показать ещё" → следующая страница
- [ ] Inline query поддержка

### Phase 4: Обогащение и рост (месяцы 2-3)
- [ ] Голосовой поиск (Web Speech API)
- [ ] Персонализация (история запросов → приоритеты)
- [ ] Мультиязычность (английский для туристов)
- [ ] Обратная связь (кнопки "Полезно"/"Бесполезно")
- [ ] Автоматическое обогащение тегами (название → terrace, отзывы → noise)
- [ ] Админ-панель для ручного обогащения POI
- [ ] Фоновый скрипт обновления OSM (еженедельно)
- [ ] Фильтрация нецензурных запросов

### Phase 5: Масштабирование и монетизация (месяцы 4-6)
- [ ] Мобильное приложение (React Native / Flutter)
- [ ] Push-уведомления ("Сегодня в вашем кафе новый десерт")
- [ ] Партнерские интеграции (заведения платят за продвижение)
- [ ] Рекламные карточки (CPM-модель)
- [ ] Премиум-функции
- [ ] Интеграция с каршерингом/такси
- [ ] Система лояльности (бонусы за шаринг)

## Метрики успеха

| Метрика | Целевое значение | Срок |
|---------|-----------------|------|
| Время ответа API | < 500 мс (p95) | Phase 1 |
| Точность NLP | > 85% | Phase 1 |
| Пользователи в неделю | 1000+ | Phase 3 |
| Повторные визиты | > 30% | Phase 4 |
| CTR на результаты | > 40% | Phase 2 |
| Доля запросов с геолокацией | > 60% | Phase 2 |
