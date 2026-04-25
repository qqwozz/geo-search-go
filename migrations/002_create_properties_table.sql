-- Таблица недвижимости
CREATE TABLE IF NOT EXISTS properties (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    address TEXT,
    price INTEGER,
    rooms INTEGER,
    area_sqm DECIMAL(8,2),
    geom GEOGRAPHY(POINT, 4326) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Комментарии к полям (документация)
COMMENT ON COLUMN properties.geom IS 'Координаты в формате (longitude, latitude) SRID=4326';