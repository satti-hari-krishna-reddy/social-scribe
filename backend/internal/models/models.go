package models



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