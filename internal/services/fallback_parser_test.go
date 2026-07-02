package services

import (
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
		Rating:    4.5,
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

func TestRankByIntent_Breakfast(t *testing.T) {
	poi := &models.POI{
		Rating:    4.0,
		DistanceMeters: 500,
		PriceLevel: 1,
		Features: models.Features{
			Breakfast: true,
		},
	}
	score := RankByIntent(poi, "breakfast")
	if score <= 0 {
		t.Errorf("expected positive score, got %f", score)
	}
}

func TestGenerateExplanation_Work(t *testing.T) {
	poi := &models.POI{
		Features: models.Features{
			Wifi:    true,
			Outlets: true,
			Quiet:   true,
		},
	}
	explanation := GenerateExplanation(poi, "work")
	if explanation == "" {
		t.Error("expected non-empty explanation")
	}
}
