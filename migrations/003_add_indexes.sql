-- GiST индекс для гео-поиска (самый важный)
CREATE INDEX IF NOT EXISTS idx_properties_geom 
ON properties USING GIST (geom);

-- Индексы для фильтрации
CREATE INDEX IF NOT EXISTS idx_properties_price 
ON properties (price);

CREATE INDEX IF NOT EXISTS idx_properties_rooms 
ON properties (rooms);