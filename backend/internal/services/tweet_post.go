package services

import (
	"errors"
	"github.com/dghubble/oauth1"
	"log"
	"net/http"
)

var twitterConfig = &oauth1.Config{}

func InitTwitterConfig(config *oauth1.Config) {
	twitterConfig = config
}

func postTweetHandler(message string, blogId string, userToken *oauth1.Token) error {

	client := twitterConfig.Client(oauth1.NoContext, userToken)

	tweetURL := "https://api.twitter.com/1.1/statuses/update.json"
	resp, err := client.PostForm(tweetURL, map[string][]string{"status": {message}})
	if err != nil {
		log.Printf("[ERROR] Failed to post tweet for the blog id : %s and the error is %s", blogId, err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("Failed to post tweet: " + resp.Status)
	}

	log.Printf("[INFO] Blog with ID %s shared on X(twitter) Successfully", blogId)
	return nil
}
