package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"geo-search/internal/services"
)

// HealthHandler godoc
// @Summary Health check
// @Description Check health of PostgreSQL, Redis, and NLP services
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 503 {object} map[string]string
// @Router /api/health [get]
func HealthHandler(pool *pgxpool.Pool, rdb *redis.Client, nlpURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		status := gin.H{"status": "ok"}

		if err := pool.Ping(c.Request.Context()); err != nil {
			status["postgres"] = "error"
			c.JSON(http.StatusServiceUnavailable, status)
			return
		}
		status["postgres"] = "ok"

		if err := rdb.Ping(c.Request.Context()).Err(); err != nil {
			status["redis"] = "error"
			c.JSON(http.StatusServiceUnavailable, status)
			return
		}
		status["redis"] = "ok"

		nlpOk := services.CheckNLPHealth(c.Request.Context(), nlpURL)
		if nlpOk {
			status["nlp"] = "ok"
		} else {
			status["nlp"] = "error"
			status["status"] = "degraded"
		}

		statusCode := http.StatusOK
		if !nlpOk {
			statusCode = http.StatusServiceUnavailable
		}
		c.JSON(statusCode, status)
	}
}
