package repositories

import (
	"context"
	"log"
	"social-scribe/backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func InsertUser(user models.User) (string, error) {
	ctx := context.TODO()
	
	result, err := userCollection.InsertOne(ctx, user)
	if err != nil {
		log.Printf("[ERROR] Error inserting user: %v", err) 
		return "", err 
	}
	id := result.InsertedID.(primitive.ObjectID).Hex() 
	return id, nil 
}

func UpdateUser(userID string, updatedUser *models.User) error {
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


func GetUserById(userID string) (*models.User, error) {
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

func GetUserByName(userName string) (*models.User, error) {
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


