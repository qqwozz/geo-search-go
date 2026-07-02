package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func CacheKey(query string, lat, lon float64, radius int) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%s:%.6f:%.6f:%d", query, lat, lon, radius)))
	return "search:" + hex.EncodeToString(h.Sum(nil))
}

func GetCache(ctx context.Context, client *redis.Client, key string) ([]byte, error) {
	return client.Get(ctx, key).Bytes()
}

func SetCache(ctx context.Context, client *redis.Client, key string, data []byte, ttl time.Duration) error {
	return client.Set(ctx, key, data, ttl).Err()
}
