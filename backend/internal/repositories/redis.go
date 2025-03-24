// repositories/redis.go
package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client
var IsRateLimited = defaultIsRateLimited

// InitRedis initializes a persistent connection to Redis.
func InitRedis() {
	redisAddr := os.Getenv("REDIS_ADDR")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisDB, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		log.Println("[WARN] REDIS_DB not set or invalid, using default value")
		redisDB = 0
	}
	if redisAddr == "" {
		log.Println("[WARN] REDIS_ADDR not set, using default value")
		redisAddr = "localhost:6379"
	}
	if redisPassword == "" {
		log.Println("[WARN] REDIS_PASSWORD not set, trying to connect without password")
	}

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := RedisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("[ERROR] Failed connecting to Redis: %v", err)
	}
	log.Println("[INFO] Successfully connected to Redis")
}

func SetRcache(key string, value interface{}, expiration time.Duration) error {
	ctx := context.Background()

	jsonData, err := json.Marshal(value)
	if err != nil {
		log.Printf("[ERROR] Error marshalling value for key %s: %v", key, err)
		return err
	}

	if err := RedisClient.Set(ctx, key, jsonData, expiration).Err(); err != nil {
		log.Printf("[ERROR] Error setting cache for key %s: %v", key, err)
		return err
	}

	return nil
}

func GetRcache(key string) (interface{}, bool) {
	ctx := context.Background()

	result, err := RedisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			// Key does not exist
			return nil, false
		}
		log.Printf("[ERROR] Error getting cache for key %s: %v", key, err)
		return nil, false
	}

	var value interface{}
	if err := json.Unmarshal([]byte(result), &value); err != nil {
		log.Printf("[ERROR] Error unmarshalling value for key %s: %v", key, err)
		return nil, false
	}

	return value, true
}

func DeleteRcache(key string) error {
	ctx := context.Background()

	if err := RedisClient.Del(ctx, key).Err(); err != nil {
		log.Printf("[ERROR] Error deleting cache for key %s: %v", key, err)
		return err
	}

	return nil
}

func defaultIsRateLimited(userID, path string, limit int, duration time.Duration) bool {
	ctx := context.Background()
	key := fmt.Sprintf("rate_limit:%s:%s", userID, path)

	count, err := RedisClient.Incr(ctx, key).Result()
	if err != nil {
		log.Printf("[ERROR] Redis INCR error: %v", err)
		return false
	}

	if count == 1 {
		RedisClient.Expire(ctx, key, duration)
	}

	return count > int64(limit)
}
