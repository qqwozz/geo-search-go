## Установка и запуск

### 1. Установить PostgreSQL и PostGIS

### 2. Создать базу данных

```bash
createdb -U postgres realty_db
psql -U postgres -d realty_db -c "CREATE EXTENSION postgis;"
```

### 3. Применить миграции

**bash**

```
psql -U postgres -d realty_db -f migrations/002_create_properties_table.sql
psql -U postgres -d realty_db -f migrations/003_add_indexes.sql
```

### 4. (Опционально) Загрузить тестовые данные

**bash**

```
psql -U postgres -d realty_db -f migrations/seed.sql
```

**text**

```

---

## 🚀 Продвинутый вариант: Использовать миграции в Go коде

Чтобы при запуске Go-приложения БД автоматически создавалась и обновлялась.

### Установите библиотеку для миграций:

```bash
go get -u github.com/golang-migrate/migrate/v4
go get -u github.com/golang-migrate/migrate/v4/database/postgres
go get -u github.com/golang-migrate/migrate/v4/source/file
```

### Миграции (файлы в папке `migrations/`):

**text**

```
migrations/
├── 001_create_properties_table.up.sql
├── 001_create_properties_table.down.sql
├── 002_add_indexes.up.sql
├── 002_add_indexes.down.sql
```

### `001_create_properties_table.up.sql`:

**sql**

```
CREATE TABLE IF NOT EXISTS properties (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    address TEXT,
    price INTEGER,
    rooms INTEGER,
    area_sqm DECIMAL(8,2),
    geom GEOGRAPHY(POINT, 4326) NOT NULL
);
```

### `001_create_properties_table.down.sql`:

**sql**

```
DROP TABLE IF EXISTS properties;
```

### `main.go` с автоматическими миграциями:

**go**

```
package main

import (
    "log"
    "github.com/golang-migrate/migrate/v4"
    _ "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
    // Запуск миграций при старте приложения
    m, err := migrate.New(
        "file://migrations",                              // папка с миграциями
        "postgres://postgres:password@localhost:5432/realty_db?sslmode=disable",
    )
    if err != nil {
        log.Fatal(err)
    }

    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        log.Fatal(err)
    }

    log.Println("Миграции применены успешно")

    // Запуск вашего API сервера...
}
```
