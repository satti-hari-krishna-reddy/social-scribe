package models

import (
    "go.mongodb.org/mongo-driver/bson/primitive"
    "time"
)

type User struct {
    Id                 primitive.ObjectID `bson:"_id,omitempty"`
    UserName           string             `bson:"username"`
    PassWord           string             `bson:"password"`
    Verified           bool               `bson:"verified"`
    WebSocketUrl       string             `bson:"websocket_url"`
    XoauthKey          string             `bson:"xoauth_key"`
    LinkedInOauthKey   string             `bson:"linkedin_oauth_key"`
    HashnodePAT        string             `bson:"hashnode_pat"`
    SharedBlogs        []SharedBlog       `bson:"shared_posts"`
    ScheduledBlogs     []ScheduledBlog    `bson:"scheduled_posts"`
    Notifications      []string           `bson:"notifications"`
}

type Session struct {
    PartitionKey string    `bson:"partition_key"`
    RowKey       string    `bson:"row_key"`
    UserID       string    `bson:"user_id"`
    LastActive   time.Time `bson:"last_active"`
    ExpiresAt    time.Time `bson:"expires_at"`
}


type LoginStruct struct {
	Username string `bson:"username"`
	Password string `bson:"password"`
}

type Blog struct {
	Id        string
	Title     string
	Brief     string
	Url       string
	Date      string
	Platforms []string
}

type SharedBlog struct {
	Blog
}

type ScheduledBlog struct {
	Blog
}