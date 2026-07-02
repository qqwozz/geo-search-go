# Roadmap: Geo Search B2C

## Визия
B2C умный геопоисковик для жителей города. Пользователь пишет на естественном языке, получает персонализированные рекомендации мест с объяснениями.

## Архитектура

```
Frontend (React/Vite + Leaflet) -> Go Backend (Gin) -> Python NLP (FastAPI)
                                        |
                               PostgreSQL+PostGIS + Redis
```

---

## Phase 1: Core Backend (недели 1-3)

**Цель:** Рабочий Go API + Python NLP + БД + Docker

### Задачи
- [x] Замена схемы БД: properties -> pois
- [x] Go бэкенд: /search, /autocomplete, /health
- [x] Python NLP: парсинг русского текста -> JSON фильтры
- [x] Fallback regex-парсер в Go
- [x] Intent-based ранжирование (work, breakfast, romantic, default)
- [x] Генерация объяснений к результатам
- [x] Redis кеширование (TTL 5 мин)
- [x] Rate limiting (30 req/min)
- [x] OSM ингест (cafes, restaurants, bars)
- [x] Docker Compose (postgres+postgis, redis, api, nlp)
- [x] Тесты Go + Python

### Структура файлов
```
cmd/api/main.go                    -- Go сервер (Gin)
internal/config/config.go          -- Конфиг из env
internal/models/poi.go             -- POI типы
internal/database/postgres.go      -- pgx pool
internal/database/redis.go         -- Redis клиент
internal/handlers/search.go        -- /search
internal/handlers/autocomplete.go  -- /autocomplete
internal/handlers/health.go        -- /health
internal/services/search.go        -- Оркестратор поиска
internal/services/nlp.go           -- HTTP клиент к NLP
internal/services/cache.go         -- Redis кеш
internal/services/fallback_parser.go -- Regex парсер
internal/services/ranker.go        -- Ранжирование
internal/services/explainer.go     -- Генерация объяснений
nlp/main.py                        -- FastAPI
nlp/parser.py                      -- NLP логика
nlp/dictionaries.py                -- Словари
nlp/location_parser.py             -- Парсер локации
scripts/ingest_osm.py              -- OSM ингест
docker-compose.yml                 -- Оркестрация
Dockerfile                         -- Go контейнер
nlp/Dockerfile                     -- Python контейнер
```

### Верификация
```bash
docker compose up -d
docker compose exec api python scripts/ingest_osm.py
curl "http://localhost:8080/api/search?q=тихое+кафе+с+террасой&lat=55.75&lon=37.62&radius=1000"
go test ./internal/services/...
cd nlp && pytest tests/
```

---

## Phase 2: Web Frontend (неделя 4)

**Цель:** Веб-интерфейс с картой и поиском

### Задачи
- [ ] React + Vite
- [ ] Поисковая строка с автокомплитом
- [ ] Чипы быстрых фильтров (Кофе, Завтрак, Работа, Терраса, Ужин)
- [ ] Карта Leaflet/OpenStreetMap с маркерами
- [ ] Карточки результатов (рейтинг, расстояние, теги, объяснение)
- [ ] Модалка детальной информации (адрес, телефон, часы, маршрут)
- [ ] Кнопка "Поделиться" (URL с query params)
- [ ] Избранное + история (LocalStorage)
- [ ] Геолокация через браузерный API
- [ ] Nginx в Docker для production

### Структура файлов
```
frontend/
  src/
    App.jsx
    components/
      SearchBar.jsx        -- Поиск + автокомплит
      QuickFilters.jsx     -- Чипы фильтров
      ResultsList.jsx      -- Список карточек
      POICard.jsx          -- Карточка места
      MapView.jsx          -- Leaflet карта
      DetailModal.jsx      -- Детали места
      Header.jsx
    hooks/
      useSearch.js
      useGeolocation.js
      useAutocomplete.js
    utils/
      api.js
      storage.js           -- LocalStorage
  Dockerfile
  nginx.conf
```

### Верификация
```bash
cd frontend && npm run dev
# Открыть http://localhost:3000
# Ввести "кафе с террасой" -> проверить автокомплит
# Найти места -> проверить карту и карточки
# Кликнуть на место -> проверить модалку
# Нажать "Поделиться" -> проверить URL
# Сохранить в избранное -> обновить страницу -> проверить
```

---

## Phase 3: Telegram Bot (неделя 5)

**Цель:** Telegram-бот для мобильных пользователей

### Задачи
- [ ] python-telegram-bot
- [ ] /start, /help команды
- [ ] Текстовый запрос -> топ-5 результатов с кнопками
- [ ] "Показать на карте" -> ссылка на Яндекс.Карты
- [ ] "Маршрут" -> навигация
- [ ] "Сохранить" -> избранное в сессии
- [ ] "Показать еще" -> следующая страница
- [ ] Inline query поддержка

### Структура файлов
```
bot/
  main.py
  handlers/
    start.py
    search.py
    inline.py
  keyboards.py
  config.py
  Dockerfile
```

### Верификация
```bash
docker compose up bot
# Написать боту: "кафе с вайфаем рядом с Парк Культуры"
# Проверить топ-5 результатов с кнопками
# Нажать "Показать на карте" -> проверить ссылку
# Нажать "Показать еще" -> проверить следующую страницу
```

---

## Phase 4: Обогащение и рост (месяцы 2-3)

### Задачи
- [ ] Голосовой поиск (Web Speech API)
- [ ] Персонализация (история запросов -> приоритеты)
- [ ] Мультиязычность (английский для туристов)
- [ ] Обратная связь (кнопки "Полезно"/"Бесполезно")
- [ ] Автоматическое обогащение тегами (название -> terrace, отзывы -> noise)
- [ ] Админ-панель для ручного обогащения POI
- [ ] Фоновый скрипт обновления OSM (еженедельно)
- [ ] Фильтрация нецензурных запросов

---

## Phase 5: Масштабирование и монетизация (месяцы 4-6)

### Задачи
- [ ] Мобильное приложение (React Native / Flutter)
- [ ] Push-уведомления ("Сегодня в вашем кафе новый десерт")
- [ ] Партнерские интеграции (заведения платят за продвижение)
- [ ] Рекламные карточки (CPM-модель)
- [ ] Премиум-функции
- [ ] Интеграция с каршерингом/такси
- [ ] Система лояльности (бонусы за шаринг)

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

| Компонент | Технология |
|-----------|------------|
| Бэкенд | Go 1.24, Gin, pgx/v5 |
| NLP | Python 3.12, FastAPI, Pydantic |
| БД | PostgreSQL 16 + PostGIS 3.4 |
| Кеш | Redis 7 |
| Фронтенд | React 18, Vite, Leaflet |
| Бот | python-telegram-bot |
| Контейнеры | Docker, Docker Compose |
| CI/CD | GitHub Actions |
