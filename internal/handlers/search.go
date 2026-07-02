package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"geo-search/internal/models"
	"geo-search/internal/services"
)

func SearchHandler(pool *pgxpool.Pool, rdb *redis.Client, nlpURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := strings.TrimSpace(c.Query("q"))
		if query == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'q' is required"})
			return
		}
		if len(query) > 500 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "query too long (max 500 chars)"})
			return
		}

		lat, err := strconv.ParseFloat(c.Query("lat"), 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "lat is required"})
			return
		}
		lon, err := strconv.ParseFloat(c.Query("lon"), 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "lon is required"})
			return
		}

		radius, _ := strconv.Atoi(c.DefaultQuery("radius", "2000"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		sort := c.DefaultQuery("sort", "relevance")

		req := &models.SearchRequest{
			Query:  query,
			Lat:    lat,
			Lon:    lon,
			Radius: radius,
			Limit:  limit,
			Sort:   sort,
		}

		resp, err := services.Search(c.Request.Context(), pool, rdb, nlpURL, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}
