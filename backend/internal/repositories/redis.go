// repositories/redis.go
package repositories

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

// InitRedis initializes a persistent connection to Redis.
func InitRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", 
		Password: "",              
		DB:       0,           
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

func IsRateLimited(userID string, limit int, duration time.Duration) bool {
	ctx := context.Background()
	key := "rate_limit:" + userID

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