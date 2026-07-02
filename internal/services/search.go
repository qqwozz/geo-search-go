package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"slices"
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

	key := CacheKey(req.Query, req.Lat, req.Lon, req.Radius)
	if cached, err := GetCache(ctx, rdb, key); err == nil {
		var resp models.SearchResponse
		if json.Unmarshal(cached, &resp) == nil {
			resp.Cached = true
			return &resp, nil
		}
	}

	nlpResp, err := ParseQuery(ctx, nlpURL, req.Query, "moscow")
	if err != nil {
		log.Printf("NLP failed, using fallback: %v", err)
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

func queryPOIs(ctx context.Context, pool *pgxpool.Pool, req *models.SearchRequest, nlp *models.NLPResponse) ([]models.POI, error) {
	query := `
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
	`
	args := []interface{}{req.Lon, req.Lat, req.Radius}
	argIdx := 4

	if nlp.Category != "" {
		query += fmt.Sprintf(" AND category = $%d", argIdx)
		args = append(args, nlp.Category)
		argIdx++
	}

	if nlp.Features["wifi"] {
		query += fmt.Sprintf(" AND has_wifi = TRUE")
	}
	if nlp.Features["terrace"] {
		query += fmt.Sprintf(" AND has_terrace = TRUE")
	}
	if nlp.Features["quiet"] {
		query += fmt.Sprintf(" AND is_quiet = TRUE")
	}
	if nlp.Features["breakfast"] {
		query += fmt.Sprintf(" AND has_breakfast = TRUE")
	}
	if nlp.Features["outlets"] {
		query += fmt.Sprintf(" AND has_outlets = TRUE")
	}
	if nlp.Features["parking"] {
		query += fmt.Sprintf(" AND has_parking = TRUE")
	}
	if nlp.Features["romantic"] {
		query += fmt.Sprintf(" AND is_romantic = TRUE")
	}
	if nlp.Features["family"] {
		query += fmt.Sprintf(" AND is_family_friendly = TRUE")
	}
	if nlp.Features["live_music"] {
		query += fmt.Sprintf(" AND has_live_music = TRUE")
	}

	switch req.Sort {
	case "distance":
		query += " ORDER BY geom <-> ST_MakePoint($1, $2)::geography"
	case "rating":
		query += " ORDER BY rating DESC"
	default:
		query += " ORDER BY geom <-> ST_MakePoint($1, $2)::geography"
	}

	query += fmt.Sprintf(" LIMIT $%d", argIdx)
	args = append(args, req.Limit)

	rows, err := pool.Query(ctx, query, args...)
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
			log.Printf("scan error: %v", err)
			continue
		}
		p.OpeningHours = hours
		pois = append(pois, p)
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
