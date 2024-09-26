package models

type User struct {
	Id                 string
	UserName           string
	PassWord           string
	Session            string
	ApiKey             string
	Verified           bool
	WebSocketUrl       string
	XoauthKey          string
	LinkedInOauthKey   string
	HashnodePAT        string
	SharedPosts        []SharedPost
	ScheduledPosts     [] ScheduledPost
	Notifications      []string
}

type SharedPost struct {
	Id                  string
	Title               string
	Brief               string
	Url                 string
	Date                string
	Platforms           []string
}

type ScheduledPost struct {
	Id                  string
	Title               string
	Brief               string
	Url                 string
	Date                string
	Platforms           []string
}