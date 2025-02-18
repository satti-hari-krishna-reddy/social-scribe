package repositories

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"social-scribe/backend/internal/models"
)

var GetCache = defualtGetCache
var SetCache = defualtSetCache
var DeleteCache = defualtDeleteCache

func defualtSetCache(key string, value interface{}, expiration time.Duration) error {
	ctx := context.TODO()

	item := models.CacheItem{
		Key:   key,
		Value: value,
	}

	if expiration > 0 {
		item.ExpiresAt = time.Now().Add(expiration)
	}

	_, err := cacheCollection.UpdateOne(
		ctx,
		bson.M{"key": key},
		bson.M{"$set": item},
		options.Update().SetUpsert(true),
	)

	if err != nil {
		log.Printf("[ERROR] Error setting cache for key %s: %v", key, err)
		return err
	}

	return nil
}

func defualtGetCache(key string) (interface{}, bool) {
	ctx := context.TODO()

	var result models.CacheItem
	err := cacheCollection.FindOne(ctx, bson.M{"key": key}).Decode(&result)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, false
		}
		log.Printf("[ERROR] Error getting cache for key %s: %v", key, err)
		return nil, false
	}

	// Double-check expiration in case TTL cleanup hasn't happened yet
	if !result.ExpiresAt.IsZero() && time.Now().After(result.ExpiresAt) {
		DeleteCache(key)
		return nil, false
	}

	return result, true
}

func defualtDeleteCache(key string) error {
	ctx := context.TODO()

	_, err := cacheCollection.DeleteOne(ctx, bson.M{"key": key})
	if err != nil {
		log.Printf("[ERROR] Error deleting cache for key %s: %v", key, err)
	}
	return err
}
