package services

import "geo-search/internal/models"

func RankByIntent(poi *models.POI, intent string) float64 {
	var score float64

	switch intent {
	case "work":
		if poi.Features.Wifi {
			score += 3
		}
		if poi.Features.Outlets {
			score += 3
		}
		if poi.Features.Quiet {
			score += 2
		}
		score += (poi.Rating / 5.0) * 2
		if poi.DistanceMeters > 0 {
			score += (1.0 / poi.DistanceMeters) * 100
		}
		if poi.Features.LiveMusic {
			score -= 3
		}

	case "breakfast":
		if poi.Features.Breakfast {
			score += 5
		}
		score += (poi.Rating / 5.0) * 3
		if poi.PriceLevel <= 2 {
			score += 2
		}
		if poi.DistanceMeters > 0 {
			score += (1.0 / poi.DistanceMeters) * 50
		}

	case "romantic":
		if poi.Features.Romantic {
			score += 4
		}
		score += (poi.Rating / 5.0) * 2
		if poi.Features.Quiet {
			score += 2
		}
		if poi.DistanceMeters > 0 {
			score += (1.0 / poi.DistanceMeters) * 30
		}

	default:
		score += (poi.Rating / 5.0) * 5
		if poi.DistanceMeters > 0 {
			score += (1.0 / poi.DistanceMeters) * 50
		}
		if poi.ReviewCount > 0 {
			score += (float64(poi.ReviewCount) / 100.0) * 2
		}
	}

	return score
}
