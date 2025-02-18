package services

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

var (
	mailgunAPIKey string
	domain        string
	senderName    string
	senderEmail   string
	SendEmail     = defualtSendEmail
	GenerateOTP   = defaultGenerateOTP
)

func init() {
	if os.Getenv("TEST_ENV") == "true" {
		return // Skip loading .env in tests ?? Hmmm, is there a beter way ?
	}
	envPath := os.Getenv("ENV_PATH")
	if envPath == "" {
		envPath = "../../.env"
	}
	if err := godotenv.Load(envPath); err != nil {
		log.Fatalf("Error loading .env file from %s: %v", envPath, err)
	}

	mailgunAPIKey = os.Getenv("MAILGUN_API_KEY")
	domain = os.Getenv("MAILGUN_DOMAIN")
	senderName = os.Getenv("MAILGUN_SENDER_NAME")
	senderEmail = os.Getenv("MAILGUN_SENDER_EMAIL")
}

func defaultGenerateOTP() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

func defualtSendEmail(toEmail, message string) error {
	mailgunURL := fmt.Sprintf("https://api.mailgun.net/v3/%s/messages", domain)

	data := url.Values{}
	data.Set("from", fmt.Sprintf("%s <%s>", senderName, senderEmail))
	data.Set("to", toEmail)
	data.Set("subject", "Your OTP Code for Social Scribe")
	data.Set("text", message)

	req, err := http.NewRequest("POST", mailgunURL, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.SetBasicAuth("api", mailgunAPIKey)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}

	const maxRetries = 3
	var resp *http.Response

	for i := 0; i < maxRetries; i++ {
		resp, err = client.Do(req)
		if err == nil && (resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusAccepted) {
			break
		}
		if resp != nil {
			resp.Body.Close()
		}
		waitTime := time.Duration((i+1)*2) * time.Second
		time.Sleep(waitTime)
	}

	if err != nil {
		return fmt.Errorf("failed to send email after %d attempts, last error: %v", maxRetries, err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("failed to send email, status code: %d", resp.StatusCode)
	}

	defer resp.Body.Close()
	return nil
}
