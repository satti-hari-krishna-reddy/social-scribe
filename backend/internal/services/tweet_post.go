package services

import (
	"bytes"
	"encoding/json"
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
	// Trim tweet if it exceeds 280 chars
	runes := []rune(message)
	if len(runes) > 280 {
		log.Printf("[WARN] Tweet for blog id %s exceeds 280 characters, trimming message", blogId)
		message = string(runes[:277]) + "..."
	}

	client := twitterConfig.Client(oauth1.NoContext, userToken)
	tweetURL := "https://api.twitter.com/2/tweets"

	// Create JSON payload
	payload, err := json.Marshal(map[string]interface{}{
		"text": message,
	})
	if err != nil {
		log.Printf("[ERROR] Failed to marshal tweet payload for blog id %s: %s", blogId, err)
		return err
	}

	req, err := http.NewRequest("POST", tweetURL, bytes.NewBuffer(payload))
	if err != nil {
		log.Printf("[ERROR] Failed to create request for blog id %s: %s", blogId, err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	// OAuth1 signing
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[ERROR] Failed to post tweet for blog id %s: %s", blogId, err)
		return err
	}
	defer resp.Body.Close()

	// Handle response
	if resp.StatusCode != http.StatusCreated {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		log.Printf("[ERROR] Twitter API response: %v", errResp)
		return errors.New("failed to post tweet: " + resp.Status)
	}

	log.Printf("[INFO] Blog with ID %s shared on X(Twitter) successfully", blogId)
	return nil
}
