-- Для разработки и тестов
INSERT INTO properties (name, address, price, rooms, area_sqm, geom) VALUES
    ('ЖК Солнечный', 'ул. Тверская 15', 8500000, 2, 45.5, 
     ST_MakePoint(37.6185, 55.7565)::geography),
    ('Дом у Парка', 'ул. Моховая 5', 15000000, 3, 78.2, 
     ST_MakePoint(37.6170, 55.7580)::geography);