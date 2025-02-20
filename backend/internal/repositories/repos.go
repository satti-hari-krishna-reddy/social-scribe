package repositories

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client
var userCollection *mongo.Collection
var cacheCollection *mongo.Collection
var scheduledItemsCollection *mongo.Collection

func InitMongoDb() {
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		log.Println("[WARN] MONGO_URI not set, using default value")
		mongoURI = "mongodb://localhost:27017"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dbName := "social-scribe"

	clientOptions := options.Client().ApplyURI(mongoURI)
	var err error
	client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal("[ERROR] Failed connecting to MongoDB:", err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("Could not ping MongoDB:", err)
	}

	userCollection = client.Database(dbName).Collection("users")
	cacheCollection = client.Database(dbName).Collection("cache")
	scheduledItemsCollection = client.Database(dbName).Collection("scheduled_items")

	err = CreateIndexes()
	if err != nil {
		log.Println("[ERROR] Failed creating indexes:", err)
	}
	log.Println("[INFO] Successfully connected to MongoDB")
}

func CreateIndexes() error {
	ctx := context.TODO()

	indexes := []mongo.IndexModel{
		// TTL index for automatic expiration
		{
			Keys:    bson.D{{Key: "expiresAt", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(0),
		},
		// Unique index for cache keys
		{
			Keys:    bson.D{{Key: "key", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}

	_, err := cacheCollection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		log.Printf("[ERROR] Error creating indexes: %v", err)
		return err
	}

	log.Println("[INFO] Successfully created indexes for cache collection")
	return nil
}
