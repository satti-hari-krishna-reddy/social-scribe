package services

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"social-scribe/backend/internal/repositories"
	"social-scribe/backend/internal/utils"
	"time"
)

var IsIPRateLimited = defaultIsIPRateLimited

// IsIPRateLimited applies rate limiting per IP
func defaultIsIPRateLimited(r *http.Request, path string, limit int, duration time.Duration) bool {
	ctx := context.Background()
	clientIP := utils.GetClientIP(r)

	// key := "rate_limit:ip:" + clientIP
	key := fmt.Sprintf("rate_limit:ip_%s:%s", path, clientIP)

	count, err := repositories.RedisClient.Incr(ctx, key).Result()
	if err != nil {
		log.Printf("[ERROR] Redis INCR error: %v", err)
		return false
	}

	if count == 1 {
		repositories.RedisClient.Expire(ctx, key, duration)
	}

	return count > int64(limit)
}
