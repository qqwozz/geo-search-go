package main

import (
    "context"
    "fmt"
    "log"
    "github.com/jackc/pgx/v5/pgxpool"
)

func main() {
    connString := "postgres://postgres:postgres@localhost:5432/realty_db?sslmode=disable"
    
    pool, err := pgxpool.New(context.Background(), connString)
    if err != nil {
        log.Fatal("Ошибка подключения:", err)
    }
    defer pool.Close()
    
    var count int
    err = pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM properties").Scan(&count)
    if err != nil {
        log.Fatal("Ошибка запроса:", err)
    }
    
    fmt.Printf("✅ Подключено! В таблице %d записей\n", count)
    
    // Гео-запрос
    var id int
    var name string
    var distance float64
    
    err = pool.QueryRow(context.Background(), `
        SELECT id, name, 
               ST_Distance(geom, ST_MakePoint(37.6185, 55.7565)::geography) as distance
        FROM properties 
        ORDER BY geom <-> ST_MakePoint(37.6185, 55.7565)::geography 
        LIMIT 1
    `).Scan(&id, &name, &distance)
    
    if err != nil {
        log.Fatal("Гео-запрос не работает:", err)
    }
    
    fmt.Printf("✅ Ближайший объект: %s (ID=%d) на расстоянии %.0f метров\n", name, id, distance)
}