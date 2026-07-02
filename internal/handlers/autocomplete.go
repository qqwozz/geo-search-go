package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

var popularQueries = []string{
	"кафе с террасой",
	"тихое кафе с вайфаем",
	"завтрак рядом",
	"ресторан для свидания",
	"бар с живой музыкой",
	"где поработать с ноутбуком",
	"кафе с розетками",
	"семейный ресторан",
	"быстрый обед",
	"вечерний ужин",
	"кофе навынос",
	"ресторан с видом",
	"пиццерия рядом",
	"детское кафе",
	"романтичный ужин",
	"кафе для работы",
	"где позавтракать",
	"бар рядом",
	"парк для прогулки",
	"кафе с парковкой",
}

func AutocompleteHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		q := strings.ToLower(strings.TrimSpace(c.Query("q")))
		if q == "" {
			c.JSON(http.StatusOK, gin.H{"suggestions": popularQueries[:8]})
			return
		}

		var matches []string
		for _, pq := range popularQueries {
			if strings.Contains(strings.ToLower(pq), q) {
				matches = append(matches, pq)
				if len(matches) >= 8 {
					break
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{"suggestions": matches})
	}
}
