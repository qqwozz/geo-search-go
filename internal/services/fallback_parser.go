package services

import (
	"regexp"
	"strings"

	"geo-search/internal/models"
)

var categoryKeywords = map[string][]string{
	"cafe":       {"кафе", "кофейня", "кофе", "капучино", "латте", "американо", "кофейн"},
	"restaurant": {"ресторан", "столовая", "обед", "ужин", "ресторанн"},
	"bar":        {"бар", "пивбар", "коктейль", "вино", "пивн"},
	"fast_food":  {"фастфуд", "быстрое питание", "шаурма", "пицца", "хинкальн"},
	"park":       {"парк", "сквер", "аллея", "сад"},
}

var featureKeywords = map[string][]string{
	"wifi":      {"вайфай", "вайфаем", "вайфая", "wifi", "wi-fi", "интернет"},
	"terrace":   {"террас", "веранд", "открытая площадка", "летняя площадка"},
	"quiet":     {"тихо", "тихий", "тихое", "тихого", "спокойн", "без музыки"},
	"outlets":   {"розетк", "зарядк"},
	"breakfast": {"завтрак", "завтраки", "утреннее меню", "брэнч", "бранч"},
	"parking":   {"парковк", "паркинг", "стоянка"},
	"romantic":  {"романтич", "романтическ", "романтично", "свидани", "для двоих"},
	"family":    {"с детьми", "детская", "семейн"},
	"live_music": {"живая музыка", "живой звук", "живой музык", "концерт"},
}

var intentKeywords = map[string][]string{
	"work":      {"работ", "ноутбук", "ноут", "лаптоп", "удобно работать", "на часах"},
	"breakfast": {"завтрак"},
	"dinner":    {"ужин"},
	"romantic":  {"романтич", "свидани", "красив"},
}

var distanceKeywords = []struct {
	word string
	dist int
}{
	{"недалеко", 1000},
	{"рядом", 500},
	{"близко", 500},
	{"вблизи", 500},
	{"около", 1000},
	{"далеко", 5000},
}

var metroPattern = regexp.MustCompile(`метро\s+([А-Яа-яёЁ][А-Яа-яёЁ\s]*?)(?:\s*[,.!]|\s*$)`)

func FallbackParse(text string) *models.NLPResponse {
	lower := strings.ToLower(text)

	result := &models.NLPResponse{
		Category: "cafe",
		Intent:   "default",
		Features: make(map[string]bool),
		Radius:   2000,
		SortBy:   "relevance",
	}

	for cat, keywords := range categoryKeywords {
		for _, kw := range keywords {
			if strings.Contains(lower, kw) {
				result.Category = cat
				break
			}
		}
	}

	for feat, keywords := range featureKeywords {
		for _, kw := range keywords {
			if strings.Contains(lower, kw) {
				result.Features[feat] = true
				break
			}
		}
	}

	for intent, keywords := range intentKeywords {
		for _, kw := range keywords {
			if strings.Contains(lower, kw) {
				result.Intent = intent
				break
			}
		}
	}

	for _, dk := range distanceKeywords {
		if strings.Contains(lower, dk.word) {
			result.Radius = dk.dist
			result.RadiusRaw = dk.word
			break
		}
	}

	if matches := metroPattern.FindStringSubmatch(text); len(matches) > 1 {
		metro := strings.TrimSpace(matches[1])
		if metro != "" {
			result.Location = &models.LocationHint{Metro: metro}
		}
	}

	return result
}
