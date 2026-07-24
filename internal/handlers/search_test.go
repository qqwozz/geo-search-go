package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/search", func(c *gin.Context) {
		query := c.Query("q")
		if query == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'q' is required"})
			return
		}

		lat := c.Query("lat")
		lon := c.Query("lon")
		if lat == "" || lon == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "lat and lon are required"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"pois":  []interface{}{},
			"total": 0,
			"query": query,
		})
	})

	r.GET("/api/autocomplete", func(c *gin.Context) {
		q := c.Query("q")
		c.JSON(http.StatusOK, gin.H{"suggestions": []string{q + " cafe", q + " restaurant"}})
	})

	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return r
}

func TestSearchHandler_MissingQuery(t *testing.T) {
	router := setupRouter()

	req := httptest.NewRequest("GET", "/api/search?lat=55.755&lon=37.615", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "query parameter 'q' is required" {
		t.Errorf("unexpected error: %s", resp["error"])
	}
}

func TestSearchHandler_MissingCoordinates(t *testing.T) {
	router := setupRouter()

	req := httptest.NewRequest("GET", "/api/search?q=кафе", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestSearchHandler_Success(t *testing.T) {
	router := setupRouter()

	req := httptest.NewRequest("GET", "/api/search?q=кафе&lat=55.755&lon=37.615", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["query"] != "кафе" {
		t.Errorf("expected query 'кафе', got '%s'", resp["query"])
	}
}

func TestAutocompleteHandler(t *testing.T) {
	router := setupRouter()

	req := httptest.NewRequest("GET", "/api/autocomplete?q=кафе", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	suggestions := resp["suggestions"].([]interface{})
	if len(suggestions) != 2 {
		t.Errorf("expected 2 suggestions, got %d", len(suggestions))
	}
}

func TestHealthHandler(t *testing.T) {
	router := setupRouter()

	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["status"] != "ok" {
		t.Errorf("expected status 'ok', got '%s'", resp["status"])
	}
}
