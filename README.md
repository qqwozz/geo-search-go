# geo-search-go

API для поиска точек в радиусе, polygon-фильтрации, и изохрон (зон досягаемости).

Реальный кейс - недвижимость, каршеринг

# Geo Search API — поиск ближайшей недвижимости

API для поиска объектов недвижимости в радиусе от заданной точки с использованием **PostgreSQL + PostGIS** и **Golang**.

## 🚀 Возможности

- Поиск объектов в заданном радиусе (метры)
- Сортировка по расстоянию (от ближайших к дальним)
- Фильтрация по цене, комнатам, площади
- Возврат расстояния в метрах до каждого объекта
- Быстрый поиск через GiST индекс PostGIS

## 🛠 Технологии

| Компонент          | Технология |
| --------------------------- | -------------------- |
| Язык                    | Go 1.23+             |
| База данных       | PostgreSQL 16+       |
| Гео-расширение | PostGIS 3.4+         |
| HTTP фреймворк     | Gin                  |
| Драйвер БД         | pgx/v5               |

## 📁 Структура проекта

.
├── cmd/
│   ├── api/main.go          # API сервер
│   └── check/main.go        # Утилита проверки БД
├── internal/
│   ├── database/database.go # Подключение к БД
│   └── handlers/handlers.go # Обработчики API
├── migrations/
│   ├── 001_create_database.sql
│   ├── 002_create_properties_table.sql
│   ├── 003_add_indexes.sql
│   └── seed.sql             # Тестовые данные
├── go.mod
├── go.sum
└── README.md

```

## 🗄 Установка и запуск

### 1. Установка PostgreSQL и PostGIS

**Ubuntu/Debian:**
```bash
sudo apt update
sudo apt install postgresql postgresql-contrib postgis postgresql-postgis
```

**macOS:**

```bash
brew install postgresql postgis
brew services start postgresql
```

### 2. Создание базы данных

```bash
sudo -u postgres psql
```

```sql
CREATE DATABASE realty_db;
\c realty_db;
CREATE EXTENSION postgis;
```

### 3. Создание таблицы и индексов

```bash
psql -U postgres -d realty_db -f migrations/002_create_properties_table.sql
psql -U postgres -d realty_db -f migrations/003_add_indexes.sql
```

### 4. Загрузка тестовых данных

```bash
psql -U postgres -d realty_db -f migrations/seed.sql
```

### 5. Настройка Go проекта

```bash
cd ~/geo-search-go/app
go mod tidy
```

### 6. Запуск API сервера

```bash
go run cmd/api/main.go
```

Сервер запустится на `http://localhost:8080`

## 📡 API Эндпоинты

### GET `/api/properties/nearby`

Поиск объектов недвижимости в радиусе.

**Параметры запроса:**

| Параметр | Тип | Обязательный | По умолчанию | Описание                                           |
| ---------------- | ------ | ------------------------ | ----------------------- | ---------------------------------------------------------- |
| `lat`          | float  | ✅ Да                  | —                      | Широта центра поиска                     |
| `lon`          | float  | ✅ Да                  | —                      | Долгота центра поиска                   |
| `radius`       | int    | ❌ Нет                | 1000                    | Радиус поиска в метрах                  |
| `limit`        | int    | ❌ Нет                | 20                      | Количество результатов                |
| `min_price`    | int    | ❌ Нет                | 0                       | Минимальная цена                            |
| `max_price`    | int    | ❌ Нет                | 0                       | Максимальная цена                          |
| `min_rooms`    | int    | ❌ Нет                | 0                       | Минимальное количество комнат   |
| `max_rooms`    | int    | ❌ Нет                | 0                       | Максимальное количество комнат |

**Пример запроса:**

```bash
curl "http://localhost:8080/api/properties/nearby?lat=55.7565&lon=37.6185&radius=500&limit=10"
```

**Пример ответа:**

```json
{
  "center": {
    "lat": 55.7565,
    "lon": 37.6185
  },
  "properties": [
    {
      "id": 1,
      "name": "ЖК Солнечный",
      "address": "ул. Тверская 15",
      "price": 8500000,
      "rooms": 2,
      "area_sqm": 45.5,
      "lat": 55.7565,
      "lon": 37.6185,
      "distance_meters": 0
    },
    {
      "id": 2,
      "name": "Дом у Парка",
      "address": "ул. Моховая 5",
      "price": 15000000,
      "rooms": 3,
      "area_sqm": 78.2,
      "lat": 55.758,
      "lon": 37.617,
      "distance_meters": 191.73
    }
  ],
  "total": 2
}
```

### GET `/api/health`

Проверка работоспособности сервера.

```bash
curl "http://localhost:8080/api/health"
```

## 🧪 Проверка базы данных

```bash
go run cmd/check/main.go
```

Ожидаемый вывод:

```
✅ Подключено! В таблице 3 записей
✅ Ближайший объект: ЖК Солнечный (ID=1) на расстоянии 0 метров
```

## 📊 Примеры SQL запросов

### Поиск в радиусе 500 метров

```sql
SELECT id, name,
       ST_Distance(geom, ST_MakePoint(37.6185, 55.7565)::geography) as distance
FROM properties
WHERE ST_DWithin(geom, ST_MakePoint(37.6185, 55.7565)::geography, 500)
ORDER BY geom <-> ST_MakePoint(37.6185, 55.7565)::geography
LIMIT 10;
```

### Фильтрация по цене и комнатам

```sql
SELECT id, name, price, rooms
FROM properties
WHERE ST_DWithin(geom, ST_MakePoint(37.6185, 55.7565)::geography, 1000)
  AND price BETWEEN 5000000 AND 15000000
  AND rooms >= 2
ORDER BY price;
```

## ⚡ Производительность

- GiST индекс на колонке `geom` обеспечивает O(log n) поиск
- `ST_DWithin` использует индекс, не сканирует всю таблицу
- Оператор `<->` сортирует по расстоянию без дополнительных вычислений
