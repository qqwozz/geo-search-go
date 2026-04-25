package main

import (
    "context"
    "log"
    "strconv"
    
    "github.com/gin-gonic/gin"
    "github.com/jackc/pgx/v5/pgxpool"
)

type Property struct {
    ID       int     `json:"id"`
    Name     string  `json:"name"`
    Address  string  `json:"address"`
    Price    int     `json:"price"`
    Rooms    int     `json:"rooms"`
    AreaSqm  float64 `json:"area_sqm"`
    Lat      float64 `json:"lat"`
    Lon      float64 `json:"lon"`
    Distance float64 `json:"distance_meters"`
}

var db *pgxpool.Pool

func main() {
    var err error
    db, err = pgxpool.New(context.Background(), "postgres://postgres:postgres@localhost:5432/realty_db?sslmode=disable")
    if err != nil {
        log.Fatal("Ошибка подключения к БД:", err)
    }
    defer db.Close()
    
    r := gin.Default()
    
    r.GET("/api/properties/nearby", searchNearby)
    r.GET("/api/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })
    
    log.Println("Сервер запущен на http://localhost:8080")
    r.Run(":8080")
}

func searchNearby(c *gin.Context) {
    lat, _ := strconv.ParseFloat(c.Query("lat"), 64)
    lon, _ := strconv.ParseFloat(c.Query("lon"), 64)
    radius, _ := strconv.Atoi(c.DefaultQuery("radius", "1000"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
    
    if lat == 0 || lon == 0 {
        c.JSON(400, gin.H{"error": "lat и lon обязательны"})
        return
    }
    
    query := `
        SELECT 
            id, name, address, price, rooms, area_sqm,
            ST_Y(geom::geometry) as lat,
            ST_X(geom::geometry) as lon,
            ST_Distance(geom, ST_MakePoint($1, $2)::geography) as distance
        FROM properties
        WHERE ST_DWithin(geom, ST_MakePoint($1, $2)::geography, $3)
        ORDER BY geom <-> ST_MakePoint($1, $2)::geography
        LIMIT $4
    `
    
    rows, err := db.Query(context.Background(), query, lon, lat, radius, limit)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    defer rows.Close()
    
    var properties []Property
    for rows.Next() {
        var p Property
        err := rows.Scan(&p.ID, &p.Name, &p.Address, &p.Price, &p.Rooms, &p.AreaSqm, &p.Lat, &p.Lon, &p.Distance)
        if err != nil {
            continue
        }
        properties = append(properties, p)
    }
    
    c.JSON(200, gin.H{
        "properties": properties,
        "total":      len(properties),
        "center":     gin.H{"lat": lat, "lon": lon},
    })
}