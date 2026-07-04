package services

import (
	"strings"
	"testing"

	"geo-search/internal/models"
)

func TestFallbackParse_Cafe(t *testing.T) {
	result := FallbackParse("кафе с террасой")
	if result.Category != "cafe" {
		t.Errorf("expected category 'cafe', got '%s'", result.Category)
	}
	if !result.Features["terrace"] {
		t.Error("expected terrace feature to be true")
	}
}

func TestFallbackParse_Restaurant(t *testing.T) {
	result := FallbackParse("ресторан для ужина")
	if result.Category != "restaurant" {
		t.Errorf("expected category 'restaurant', got '%s'", result.Category)
	}
	if result.Intent != "dinner" {
		t.Errorf("expected intent 'dinner', got '%s'", result.Intent)
	}
}

func TestFallbackParse_WorkIntent(t *testing.T) {
	result := FallbackParse("где поработать с ноутбуком")
	if result.Intent != "work" {
		t.Errorf("expected intent 'work', got '%s'", result.Intent)
	}
}

func TestFallbackParse_Distance(t *testing.T) {
	result := FallbackParse("кафе недалеко")
	if result.Radius != 1000 {
		t.Errorf("expected radius 1000, got %d", result.Radius)
	}
	if result.RadiusRaw != "недалеко" {
		t.Errorf("expected radius_raw 'недалеко', got '%s'", result.RadiusRaw)
	}
}

func TestFallbackParse_Metro(t *testing.T) {
	result := FallbackParse("кафе рядом с метро Парк Культуры")
	if result.Location == nil {
		t.Fatal("expected location to be non-nil")
	}
	if result.Location.Metro != "Парк Культуры" {
		t.Errorf("expected metro 'Парк Культуры', got '%s'", result.Location.Metro)
	}
}

func TestFallbackParse_QuietWifi(t *testing.T) {
	result := FallbackParse("тихое кафе с вайфаем")
	if !result.Features["quiet"] {
		t.Error("expected quiet feature to be true")
	}
	if !result.Features["wifi"] {
		t.Error("expected wifi feature to be true")
	}
}

func TestFallbackParse_Breakfast(t *testing.T) {
	result := FallbackParse("где позавтракать")
	if !result.Features["breakfast"] {
		t.Error("expected breakfast feature to be true")
	}
	if result.Intent != "breakfast" {
		t.Errorf("expected intent 'breakfast', got '%s'", result.Intent)
	}
}

func TestFallbackParse_Romantic(t *testing.T) {
	result := FallbackParse("романтичный ресторан для свидания")
	if result.Intent != "romantic" {
		t.Errorf("expected intent 'romantic', got '%s'", result.Intent)
	}
	if !result.Features["romantic"] {
		t.Error("expected romantic feature to be true")
	}
}

func TestFallbackParse_Bar(t *testing.T) {
	result := FallbackParse("бар с живой музыкой")
	if result.Category != "bar" {
		t.Errorf("expected category 'bar', got '%s'", result.Category)
	}
	if !result.Features["live_music"] {
		t.Error("expected live_music feature to be true")
	}
}

func TestFallbackParse_Default(t *testing.T) {
	result := FallbackParse("что-нибудь вкусное")
	if result.Category != "cafe" {
		t.Errorf("expected default category 'cafe', got '%s'", result.Category)
	}
	if result.Intent != "default" {
		t.Errorf("expected intent 'default', got '%s'", result.Intent)
	}
	if result.Radius != 2000 {
		t.Errorf("expected default radius 2000, got %d", result.Radius)
	}
}

func TestRankByIntent_Work(t *testing.T) {
	poi := &models.POI{
		Rating:         4.5,
		DistanceMeters: 200,
		Features: models.Features{
			Wifi:    true,
			Outlets: true,
			Quiet:   true,
		},
	}
	score := RankByIntent(poi, "work")
	if score <= 0 {
		t.Errorf("expected positive score, got %f", score)
	}
}

func TestRankByIntent_WorkPenalizesLiveMusic(t *testing.T) {
	poiWith := &models.POI{
		Rating:         4.5,
		DistanceMeters: 200,
		Features:       models.Features{Wifi: true, Outlets: true, Quiet: true, LiveMusic: true},
	}
	poiWithout := &models.POI{
		Rating:         4.5,
		DistanceMeters: 200,
		Features:       models.Features{Wifi: true, Outlets: true, Quiet: true},
	}
	scoreWith := RankByIntent(poiWith, "work")
	scoreWithout := RankByIntent(poiWithout, "work")
	if scoreWith >= scoreWithout {
		t.Errorf("live_music should penalize work score: with=%f without=%f", scoreWith, scoreWithout)
	}
}

func TestRankByIntent_Breakfast(t *testing.T) {
	poi := &models.POI{
		Rating:         4.0,
		DistanceMeters: 500,
		PriceLevel:     1,
		Features:       models.Features{Breakfast: true},
	}
	score := RankByIntent(poi, "breakfast")
	if score <= 0 {
		t.Errorf("expected positive score, got %f", score)
	}
}

func TestRankByIntent_BreakfastPrefersCheap(t *testing.T) {
	cheap := &models.POI{Rating: 4.0, DistanceMeters: 500, PriceLevel: 1, Features: models.Features{Breakfast: true}}
	expensive := &models.POI{Rating: 4.0, DistanceMeters: 500, PriceLevel: 3, Features: models.Features{Breakfast: true}}
	if RankByIntent(cheap, "breakfast") <= RankByIntent(expensive, "breakfast") {
		t.Error("cheap place should rank higher for breakfast")
	}
}

func TestRankByIntent_Romantic(t *testing.T) {
	poi := &models.POI{
		Rating:         4.5,
		DistanceMeters: 300,
		Features:       models.Features{Romantic: true, Quiet: true},
	}
	score := RankByIntent(poi, "romantic")
	if score <= 0 {
		t.Errorf("expected positive score, got %f", score)
	}
}

func TestRankByIntent_Default(t *testing.T) {
	poi := &models.POI{
		Rating:         4.5,
		DistanceMeters: 100,
		ReviewCount:    200,
	}
	score := RankByIntent(poi, "default")
	if score <= 0 {
		t.Errorf("expected positive score, got %f", score)
	}
}

func TestRankByIntent_CloserIsBetter(t *testing.T) {
	near := &models.POI{Rating: 4.0, DistanceMeters: 100}
	far := &models.POI{Rating: 4.0, DistanceMeters: 5000}
	if RankByIntent(near, "default") <= RankByIntent(far, "default") {
		t.Error("closer place should rank higher")
	}
}

func TestRankByIntent_HigherRatingIsBetter(t *testing.T) {
	high := &models.POI{Rating: 5.0, DistanceMeters: 500}
	low := &models.POI{Rating: 2.0, DistanceMeters: 500}
	if RankByIntent(high, "default") <= RankByIntent(low, "default") {
		t.Error("higher rated place should rank higher")
	}
}

func TestGenerateExplanation_Work(t *testing.T) {
	poi := &models.POI{
		Features: models.Features{Wifi: true, Outlets: true, Quiet: true},
	}
	explanation := GenerateExplanation(poi, "work")
	if explanation == "" {
		t.Error("expected non-empty explanation")
	}
	if !strings.Contains(explanation, "Wi-Fi") {
		t.Errorf("expected Wi-Fi in explanation, got: %s", explanation)
	}
}

func TestGenerateExplanation_Breakfast(t *testing.T) {
	poi := &models.POI{
		PriceLevel: 1,
		Features:   models.Features{Breakfast: true},
	}
	explanation := GenerateExplanation(poi, "breakfast")
	if explanation == "" {
		t.Error("expected non-empty explanation")
	}
	if !strings.Contains(explanation, "Завтраки") {
		t.Errorf("expected 'Завтраки' in explanation, got: %s", explanation)
	}
}

func TestGenerateExplanation_Romantic(t *testing.T) {
	poi := &models.POI{
		Features: models.Features{Romantic: true, Quiet: true},
	}
	explanation := GenerateExplanation(poi, "romantic")
	if explanation == "" {
		t.Error("expected non-empty explanation")
	}
	if !strings.Contains(explanation, "свидания") {
		t.Errorf("expected 'свидания' in explanation, got: %s", explanation)
	}
}

func TestGenerateExplanation_Default_HighRating(t *testing.T) {
	poi := &models.POI{Rating: 4.8, ReviewCount: 100}
	explanation := GenerateExplanation(poi, "default")
	if !strings.Contains(explanation, "4.8") {
		t.Errorf("expected rating in explanation, got: %s", explanation)
	}
}

func TestGenerateExplanation_Fallback(t *testing.T) {
	poi := &models.POI{Rating: 3.0, ReviewCount: 5}
	explanation := GenerateExplanation(poi, "default")
	if explanation == "" {
		t.Error("expected non-empty explanation")
	}
}

func TestGenerateExplanation_TerraceAndParking(t *testing.T) {
	poi := &models.POI{
		Rating:   4.0,
		Features: models.Features{Terrace: true, Parking: true},
	}
	explanation := GenerateExplanation(poi, "default")
	if !strings.Contains(explanation, "терраса") {
		t.Errorf("expected 'терраса' in explanation, got: %s", explanation)
	}
	if !strings.Contains(explanation, "парковка") {
		t.Errorf("expected 'парковка' in explanation, got: %s", explanation)
	}
}

func TestCacheKey_DifferentParamsProduceDifferentKeys(t *testing.T) {
	key1 := CacheKey("кафе", 55.75, 37.60, 2000, 20, "relevance")
	key2 := CacheKey("кафе", 55.75, 37.60, 2000, 20, "rating")
	key3 := CacheKey("кафе", 55.75, 37.60, 2000, 10, "relevance")
	key4 := CacheKey("кафе", 55.75, 37.60, 1000, 20, "relevance")

	if key1 == key2 {
		t.Error("different sort should produce different keys")
	}
	if key1 == key3 {
		t.Error("different limit should produce different keys")
	}
	if key1 == key4 {
		t.Error("different radius should produce different keys")
	}
}

func TestCacheKey_SameParamsProduceSameKey(t *testing.T) {
	key1 := CacheKey("кафе", 55.75, 37.60, 2000, 20, "relevance")
	key2 := CacheKey("кафе", 55.75, 37.60, 2000, 20, "relevance")
	if key1 != key2 {
		t.Error("same params should produce same key")
	}
}
