package models

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	Id               primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	UserName         string             `json:"username" bson:"username"`
	PassWord         string             `json:"password" bson:"password"`
	Verified         bool               `json:"verified" bson:"verified"`
	EmailVerified    bool               `json:"email_verified" bson:"email_verified"`
	HashnodeVerified bool               `json:"hashnode_verified" bson:"hashnode_verified"`
	LinkedinVerified bool               `json:"linkedin_verified" bson:"linkedin_verified"`
	XVerified        bool               `json:"x_verified" bson:"x_verified"`
	WebHookUrl       string             `json:"webhook_url" bson:"webhook_url"`
	HashnodeBlog     string             `json:"hashnode_blog" bson:"hashnode_blog"`
	XOAuthToken      string             `json:"x_oauth_token" bson:"x_oauth_token"`
	XOAuthSecret     string             `json:"x_oauth_secret" bson:"x_oauth_secret"`
	LinkedInOauthKey string             `json:"linkedin_oauth_key" bson:"linkedin_oauth_key"`
	HashnodePAT      string             `json:"hashnode_pat" bson:"hashnode_pat"`
	SharedBlogs      []SharedBlog       `json:"shared_posts" bson:"shared_posts"`
	ScheduledBlogs   []ScheduledBlog    `json:"scheduled_posts" bson:"scheduled_posts"`
	Notifications    []string           `json:"notifications" bson:"notifications"`
}

type Session struct {
	PartitionKey string    `json:"partition_key" bson:"partition_key"`
	RowKey       string    `json:"row_key" bson:"row_key"`
	UserID       string    `json:"user_id" bson:"user_id"`
	LastActive   time.Time `json:"last_active" bson:"last_active"`
	ExpiresAt    time.Time `json:"expires_at" bson:"expires_at"`
}

type LoginStruct struct {
	Username string `json:"username" bson:"username"`
	Password string `json:"password" bson:"password"`
}

type ScheduledBlogData struct {
	UserID        string        `json:"user_id" bson:"user_id"`
	ScheduledBlog ScheduledBlog `json:"blog" bson:"blog"`
}

type Blog struct {
	Id                string `json:"id" bson:"id"`
	Title             string `json:"title" bson:"title"`
	Url               string `json:"url" bson:"url"`
	CoverImage        Image  `json:"coverImage" bson:"coverImage"`
	Author            Author `json:"author" bson:"author"`
	ReadTimeInMinutes int    `json:"readTimeInMinutes" bson:"readTimeInMinutes"`
}

type Image struct {
	URL string `json:"url" bson:"url"`
}

type Author struct {
	Name string `json:"name" bson:"name"`
}

type SharedBlog struct {
	Blog
	Platforms  []string `json:"platforms" bson:"platforms"`
	SharedTime string   `json:"shared_time" bson:"shared_time"`
}

type ScheduledBlog struct {
	Blog
	Platforms     []string `json:"platforms" bson:"platforms"`
	ScheduledTime string   `json:"scheduled_time" bson:"scheduled_time"`
}

type GraphQLQuery struct {
	Query string `json:"query"`
}

type CoverImage struct {
	URL string `json:"url"`
}

type PostNode struct {
	Title             string     `json:"title"`
	URL               string     `json:"url"`
	ID                string     `json:"id"`
	CoverImage        CoverImage `json:"coverImage"`
	Author            Author     `json:"author"`
	ReadTimeInMinutes int        `json:"readTimeInMinutes"`
}

type Edge struct {
	Node PostNode `json:"node"`
}

type Posts struct {
	Edges []Edge `json:"edges"`
}

type Publication struct {
	Posts Posts `json:"posts"`
}

type Data struct {
	Publication Publication `json:"publication"`
}

type GraphQLResponse struct {
	Data Data `json:"data"`
}

type TweetRequest struct {
	Tweet string `json:"tweet"`
}

type HashnodeKey struct {
	Key string `json:"key"`
}

type CacheItem struct {
	Key       string      `bson:"key"`
	Value     interface{} `bson:"value"`
	ExpiresAt time.Time   `bson:"expiresAt,omitempty"`
}

func (b *Blog) ValidateBase() error {
	if strings.TrimSpace(b.Title) == "" {
		return fmt.Errorf("title is required")
	}

	if strings.TrimSpace(b.Url) == "" || !isValidURL(b.Url) {
		return fmt.Errorf("a valid URL is required")
	}

	if strings.TrimSpace(b.Id) == "" {
		return fmt.Errorf("ID is required")
	}

	if strings.TrimSpace(b.Author.Name) == "" {
		return fmt.Errorf("author name is required")
	}

	if strings.TrimSpace(b.CoverImage.URL) == "" || !isValidURL(b.CoverImage.URL) {
		return fmt.Errorf("a valid cover image URL is required")
	}

	return nil
}

func (sb *ScheduledBlog) Validate() error {

	if err := sb.Blog.ValidateBase(); err != nil {
		return err
	}

	if len(sb.Platforms) == 0 {
		return fmt.Errorf("at least one platform is required")
	}

	scheduledTime, err := time.Parse(time.RFC3339, sb.ScheduledTime)
	if err != nil {
		return fmt.Errorf("invalid scheduled_time format, expected YYYY-MM-DD HH:mm")
	}
	currentTime := time.Now()
	diff := scheduledTime.Sub(currentTime)

	if diff > (7 * 24 * time.Hour) {
		return fmt.Errorf("scheduled time is more than 7 days from now")
	} else if diff < 0 {
		return fmt.Errorf("scheduled time is in the past")
	}

	return nil
}

func (shb *SharedBlog) Validate() error {
	if err := shb.Blog.ValidateBase(); err != nil {
		return err
	}

	if len(shb.Platforms) == 0 {
		return fmt.Errorf("at least one platform is required")
	}

	return nil
}

func isValidURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}
