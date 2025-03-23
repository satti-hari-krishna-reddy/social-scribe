package repositories

import (
	"context"
	"fmt"
	"log"
	"social-scribe/backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	InsertUser     = defaultInsertUser
	UpdateUser     = defaultUpdateUser
	GetUserById    = defaultGetUserById
	GetUserByName  = defaultGetUserByName
	DeleteUserById = defaultDeleteUserById
)

func defaultInsertUser(user models.User) (string, error) {
	ctx := context.TODO()

	result, err := userCollection.InsertOne(ctx, user)
	if err != nil {
		log.Printf("[ERROR] Error inserting user: %v", err)
		return "", err
	}

	// Try to convert the InsertedID to a primitive.ObjectID.
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		log.Printf("[INFO] Inserted user by converting into obj id with ID: %s", oid.Hex())
		return oid.Hex(), nil
	}
	// Otherwise, fall back to returning the InsertedID as a string.
	log.Printf("[INFO] Inserted user by using insertedid directly with ID: %s", result.InsertedID)
	return fmt.Sprintf("%v", result.InsertedID), nil
}

func defaultUpdateUser(userID string, updatedUser *models.User) error {
	ctx := context.TODO()
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objID}
	update := bson.M{"$set": updatedUser}

	result, err := userCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func defaultGetUserById(userID string) (*models.User, error) {
	ctx := context.TODO()

	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}
	user := &models.User{}
	err = userCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

func defaultGetUserByName(userName string) (*models.User, error) {
	ctx := context.TODO()
	user := &models.User{}
	err := userCollection.FindOne(ctx, bson.M{"username": userName}).Decode(user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return user, err
	}
	return user, nil
}

func defaultDeleteUserById(userID string) error {
	ctx := context.TODO()
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}
	filter := bson.M{"_id": objID}
	result, err := userCollection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}
