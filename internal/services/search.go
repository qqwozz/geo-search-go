package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"geo-search/internal/models"
)

func Search(ctx context.Context, pool *pgxpool.Pool, rdb *redis.Client, nlpURL string, req *models.SearchRequest) (*models.SearchResponse, error) {
	if req.Radius <= 0 {
		req.Radius = 2000
	}
	if req.Limit <= 0 || req.Limit > 50 {
		req.Limit = 20
	}

	key := CacheKey(req.Query, req.Lat, req.Lon, req.Radius, req.Limit, req.Sort)
	if cached, err := GetCache(ctx, rdb, key); err == nil {
		var resp models.SearchResponse
		if json.Unmarshal(cached, &resp) == nil {
			resp.Cached = true
			return &resp, nil
		}
	}

	nlpResp, err := ParseQuery(ctx, nlpURL, req.Query, req.City)
	if err != nil {
		slog.Warn("NLP failed, using fallback", "error", err)
		nlpResp = FallbackParse(req.Query)
	}

	pois, err := queryPOIs(ctx, pool, req, nlpResp)
	if err != nil {
		return nil, err
	}

	for i := range pois {
		pois[i].Score = RankByIntent(&pois[i], nlpResp.Intent)
		pois[i].Explanation = GenerateExplanation(&pois[i], nlpResp.Intent)
	}

	sortByScore(pois)

	resp := &models.SearchResponse{
		POIs:  pois,
		Total: len(pois),
		Center: struct {
			Lat float64 `json:"lat"`
			Lon float64 `json:"lon"`
		}{Lat: req.Lat, Lon: req.Lon},
		Query: req.Query,
	}

	if data, err := json.Marshal(resp); err == nil {
		SetCache(ctx, rdb, key, data, 5*time.Minute)
	}

	return resp, nil
}

var allowedColumns = map[string]bool{
	"category":           true,
	"has_wifi":           true,
	"has_outlets":        true,
	"has_terrace":        true,
	"has_parking":        true,
	"has_live_music":     true,
	"has_breakfast":      true,
	"is_quiet":           true,
	"is_family_friendly": true,
	"is_romantic":        true,
	"is_dog_friendly":    true,
	"noise_level":        true,
	"rating":             true,
	"review_count":       true,
	"price_level":        true,
	"city":               true,
}

func validateColumn(col string) (string, bool) {
	if allowedColumns[col] {
		return col, false
	}
	return "", true
}

func queryPOIs(ctx context.Context, pool *pgxpool.Pool, req *models.SearchRequest, nlp *models.NLPResponse) ([]models.POI, error) {
	var query strings.Builder
	query.WriteString(`
		SELECT
			id, name, name_en, category, subcategory, address, city, phone, website,
			opening_hours::text, rating, review_count, price_level,
			has_wifi, has_outlets, has_terrace, has_parking, has_live_music, has_breakfast,
			is_quiet, is_family_friendly, is_romantic, is_dog_friendly,
			noise_level,
			ST_Y(geom::geometry) as lat,
			ST_X(geom::geometry) as lon,
			ST_Distance(geom, ST_MakePoint($1, $2)::geography) as distance
		FROM pois
		WHERE ST_DWithin(geom, ST_MakePoint($1, $2)::geography, $3)
	`)

	args := []interface{}{req.Lon, req.Lat, req.Radius}

	addFilter := func(col string, value interface{}) {
		col, invalid := validateColumn(col)
		if invalid {
			return
		}
		args = append(args, value)
		query.WriteString(fmt.Sprintf(" AND %s = $%d", col, len(args)))
	}

	if nlp.Category != "" {
		addFilter("category", nlp.Category)
	}

	featureColumns := map[string]string{
		"wifi":       "has_wifi",
		"terrace":    "has_terrace",
		"quiet":      "is_quiet",
		"breakfast":  "has_breakfast",
		"outlets":    "has_outlets",
		"parking":    "has_parking",
		"romantic":   "is_romantic",
		"family":     "is_family_friendly",
		"live_music": "has_live_music",
	}
	for feat, col := range featureColumns {
		if nlp.Features[feat] {
			if _, invalid := validateColumn(col); invalid {
				continue
			}
			query.WriteString(fmt.Sprintf(" AND %s = TRUE", col))
		}
	}

	if nlp.Location != nil && nlp.Location.Street != "" {
		args = append(args, "%"+nlp.Location.Street+"%")
		query.WriteString(fmt.Sprintf(" AND address ILIKE $%d", len(args)))
	}

	switch req.Sort {
	case "rating":
		query.WriteString(" ORDER BY rating DESC")
	default:
		query.WriteString(" ORDER BY geom <-> ST_MakePoint($1, $2)::geography")
	}

	args = append(args, req.Limit)
	query.WriteString(fmt.Sprintf(" LIMIT $%d", len(args)))

	if req.Offset > 0 {
		args = append(args, req.Offset)
		query.WriteString(fmt.Sprintf(" OFFSET $%d", len(args)))
	}

	rows, err := pool.Query(ctx, query.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pois []models.POI
	for rows.Next() {
		var p models.POI
		var hours *string
		err := rows.Scan(
			&p.ID, &p.Name, &p.NameEN, &p.Category, &p.Subcategory,
			&p.Address, &p.City, &p.Phone, &p.Website,
			&hours, &p.Rating, &p.ReviewCount, &p.PriceLevel,
			&p.Features.Wifi, &p.Features.Outlets, &p.Features.Terrace,
			&p.Features.Parking, &p.Features.LiveMusic, &p.Features.Breakfast,
			&p.Features.Quiet, &p.Features.FamilyFriendly, &p.Features.Romantic,
			&p.Features.DogFriendly,
			&p.NoiseLevel,
			&p.Lat, &p.Lon, &p.DistanceMeters,
		)
		if err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}
		p.OpeningHours = hours
		pois = append(pois, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return pois, nil
}

func sortByScore(pois []models.POI) {
	slices.SortFunc(pois, func(a, b models.POI) int {
		if b.Score > a.Score {
			return 1
		}
		if b.Score < a.Score {
			return -1
		}
		return 0
	})
}
