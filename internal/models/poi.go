package models

import "time"

type POI struct {
	ID             int       `json:"id"`
	Name           string    `json:"name"`
	NameEN         *string   `json:"name_en,omitempty"`
	Category       string    `json:"category"`
	Subcategory    *string   `json:"subcategory,omitempty"`
	Address        *string   `json:"address,omitempty"`
	City           *string   `json:"city,omitempty"`
	Phone          *string   `json:"phone,omitempty"`
	Website        *string   `json:"website,omitempty"`
	OpeningHours   *string   `json:"opening_hours,omitempty"`
	Rating         float64   `json:"rating"`
	ReviewCount    int       `json:"review_count"`
	PriceLevel     int       `json:"price_level"`
	Features       Features  `json:"features"`
	NoiseLevel     string    `json:"noise_level"`
	Lat            float64   `json:"lat"`
	Lon            float64   `json:"lon"`
	DistanceMeters float64   `json:"distance_meters,omitempty"`
	Explanation    string    `json:"explanation,omitempty"`
	Score          float64   `json:"score,omitempty"`
	LastUpdated    time.Time `json:"last_updated"`
}

type Features struct {
	Wifi          bool `json:"wifi"`
	Outlets       bool `json:"outlets"`
	Terrace       bool `json:"terrace"`
	Parking       bool `json:"parking"`
	LiveMusic     bool `json:"live_music"`
	Breakfast     bool `json:"breakfast"`
	Quiet         bool `json:"quiet"`
	FamilyFriendly bool `json:"family_friendly"`
	Romantic      bool `json:"romantic"`
	DogFriendly   bool `json:"dog_friendly"`
}

type SearchRequest struct {
	Query  string  `json:"query" form:"q"`
	Lat    float64 `json:"lat" form:"lat"`
	Lon    float64 `json:"lon" form:"lon"`
	Radius int     `json:"radius" form:"radius"`
	Limit  int     `json:"limit" form:"limit"`
	Offset int     `json:"offset" form:"offset"`
	Sort   string  `json:"sort" form:"sort"`
	City   string  `json:"city" form:"city"`
}

type SearchResponse struct {
	POIs   []POI `json:"pois"`
	Total  int   `json:"total"`
	Center struct {
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	} `json:"center"`
	Query  string `json:"query"`
	Cached bool   `json:"cached"`
}

type NLPResponse struct {
	Category  string          `json:"category"`
	Intent    string          `json:"intent"`
	Features  map[string]bool `json:"features"`
	Location  *LocationHint   `json:"location"`
	Radius    int             `json:"radius"`
	RadiusRaw string          `json:"radius_raw"`
	SortBy    string          `json:"sort_by"`
}

type LocationHint struct {
	Metro  string  `json:"metro,omitempty"`
	Street string  `json:"street,omitempty"`
	Area   string  `json:"area,omitempty"`
	Lat    float64 `json:"lat,omitempty"`
	Lon    float64 `json:"lon,omitempty"`
}
