package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func HealthHandler(pool *pgxpool.Pool, rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		status := gin.H{"status": "ok"}

		if err := pool.Ping(c.Request.Context()); err != nil {
			status["postgres"] = "error: " + err.Error()
			c.JSON(http.StatusServiceUnavailable, status)
			return
		}
		status["postgres"] = "ok"

		if err := rdb.Ping(c.Request.Context()).Err(); err != nil {
			status["redis"] = "error: " + err.Error()
			c.JSON(http.StatusServiceUnavailable, status)
			return
		}
		status["redis"] = "ok"

		c.JSON(http.StatusOK, status)
	}
}
