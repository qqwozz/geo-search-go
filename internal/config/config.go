package config

import "os"

type Config struct {
	Port          string
	DatabaseURL   string
	RedisURL      string
	NLPServiceURL string
	CORSOrigin    string
}

func Load() *Config {
	return &Config{
		Port:          getEnv("PORT", "8080"),
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/geosearch?sslmode=disable"),
		RedisURL:      getEnv("REDIS_URL", "redis://localhost:6379"),
		NLPServiceURL: getEnv("NLP_SERVICE_URL", "http://localhost:8000"),
		CORSOrigin:    getEnv("CORS_ORIGIN", "http://localhost:3000"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
