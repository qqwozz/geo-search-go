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

// SearchHandler godoc
// @Summary Search for places
// @Description Find places using natural language queries in Russian
// @Tags search
// @Accept json
// @Produce json
// @Param q query string true "Search query in Russian"
// @Param lat query number true "Latitude"
// @Param lon query number true "Longitude"
// @Param radius query int false "Search radius in meters (default 2000)"
// @Param limit query int false "Max results (default 20, max 50)"
// @Param offset query int false "Pagination offset (default 0)"
// @Param sort query string false "Sort by: relevance (default) or rating"
// @Param city query string false "City name (default moscow)"
// @Success 200 {object} models.SearchResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/search [get]
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

		if lat < -90 || lat > 90 || lon < -180 || lon > 180 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "coordinates out of range: lat [-90,90], lon [-180,180]"})
			return
		}

		radius, err := strconv.Atoi(c.DefaultQuery("radius", "2000"))
		if err != nil || radius <= 0 {
			radius = 2000
		}
		limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
		if err != nil || limit <= 0 {
			limit = 20
		}
		offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
		if err != nil || offset < 0 {
			offset = 0
		}
		sort := c.DefaultQuery("sort", "relevance")
		city := strings.TrimSpace(c.DefaultQuery("city", "moscow"))

		req := &models.SearchRequest{
			Query:  query,
			Lat:    lat,
			Lon:    lon,
			Radius: radius,
			Limit:  limit,
			Offset: offset,
			Sort:   sort,
			City:   city,
		}

		resp, err := services.Search(c.Request.Context(), pool, rdb, nlpURL, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}
