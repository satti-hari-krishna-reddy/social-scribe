package services

import (
	"context"
	"log"
	"net/http"
	"social-scribe/backend/internal/repositories"
	"social-scribe/backend/internal/utils"
	"time"
)

// IsIPRateLimited applies rate limiting per IP
func IsIPRateLimited(r *http.Request, limit int, duration time.Duration) bool {
	ctx := context.Background()
	clientIP := utils.GetClientIP(r)

	key := "rate_limit:ip:" + clientIP

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
