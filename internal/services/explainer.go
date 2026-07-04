package services

import (
	"fmt"
	"strings"

	"geo-search/internal/models"
)

func GenerateExplanation(poi *models.POI, intent string) string {
	var parts []string

	switch intent {
	case "work":
		var features []string
		if poi.Features.Wifi {
			features = append(features, "быстрый Wi-Fi")
		}
		if poi.Features.Outlets {
			features = append(features, "удобные розетки")
		}
		if poi.Features.Quiet {
			features = append(features, "очень тихо")
		}
		if len(features) > 0 {
			parts = append(parts, fmt.Sprintf("Отличное место для работы: %s", strings.Join(features, ", ")))
		}

	case "breakfast":
		if poi.Features.Breakfast {
			parts = append(parts, "Завтраки есть в меню")
		}
		if poi.PriceLevel <= 2 {
			parts = append(parts, "доступные цены")
		}

	case "romantic":
		var features []string
		if poi.Features.Romantic {
			features = append(features, "романтичная атмосфера")
		}
		if poi.Features.Quiet {
			features = append(features, "тихо")
		}
		if len(features) > 0 {
			parts = append(parts, fmt.Sprintf("Идеально для свидания: %s", strings.Join(features, ", ")))
		}

	default:
		if poi.Rating >= 4.5 {
			parts = append(parts, fmt.Sprintf("Высокий рейтинг %.1f", poi.Rating))
		}
		if poi.ReviewCount > 50 {
			parts = append(parts, fmt.Sprintf("много отзывов (%d)", poi.ReviewCount))
		}
	}

	var features []string
	if poi.Features.Terrace {
		features = append(features, "терраса")
	}
	if poi.Features.Parking {
		features = append(features, "парковка")
	}
	if poi.Features.FamilyFriendly {
		features = append(features, "семейное")
	}
	if len(features) > 0 {
		parts = append(parts, fmt.Sprintf("Также есть: %s", strings.Join(features, ", ")))
	}

	if len(parts) == 0 {
		parts = append(parts, "Хороший вариант поблизости")
	}

	return strings.Join(parts, ". ")
}
