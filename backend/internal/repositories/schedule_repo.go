package repositories

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"social-scribe/backend/internal/models"
)

func GetScheduledTasks() ([]models.ScheduledBlogData, error) {
	ctx := context.TODO()

	var scheduledTasks []models.ScheduledBlogData

	cursor, err := scheduledItemsCollection.Find(ctx, bson.M{})
	if err != nil {
		log.Printf("[ERROR] Error getting scheduled tasks: %v", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &scheduledTasks); err != nil {
		log.Printf("[ERROR] Error decoding scheduled tasks: %v", err)
		return nil, err
	}

	return scheduledTasks, nil
}

func StoreScheduledTask(task models.ScheduledBlogData) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := scheduledItemsCollection.InsertOne(ctx, task)
	if err != nil {
		log.Printf("[ERROR] Failed to store scheduled task: %v", err)
		return err
	}
	return nil
}

func DeleteScheduledTask(task models.ScheduledBlogData) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Delete based on both user_id and blog.id  extra safety!
	_, err := scheduledItemsCollection.DeleteOne(ctx, bson.M{
		"user_id": task.UserID, 
		"blog.id": task.ScheduledBlog.Id,
	})
	if err != nil {
		log.Printf("[ERROR] Failed to delete scheduled task: %v", err)
		return err
	}
	return nil
}
