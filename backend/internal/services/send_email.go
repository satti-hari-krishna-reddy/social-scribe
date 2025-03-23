package services

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
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
		return
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
	senderEmail = os.Getenv("MAILGUN_EMAIL")

	if !strings.Contains(senderEmail, "@") {
		log.Fatalf("MAILGUN_EMAIL is invalid: %s", senderEmail)
	}
}

func defaultGenerateOTP() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

func defualtSendEmail(toEmail, message string) error {
	reqURL := fmt.Sprintf("https://api.mailgun.net/v3/%s/messages", domain)
	data := &bytes.Buffer{}
	writer := multipart.NewWriter(data)

	sender := fmt.Sprintf("%s <%s>", senderName, senderEmail)

	// Add fields
	_ = writer.WriteField("from", sender)
	_ = writer.WriteField("to", toEmail)
	_ = writer.WriteField("subject", "Your OTP Code for Social Scribe")
	_ = writer.WriteField("text", message)

	writer.Close()

	req, err := http.NewRequest("POST", reqURL, data)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.SetBasicAuth("api", mailgunAPIKey)
	req.Header.Add("Content-Type", writer.FormDataContentType())

	// Send request with retries
	client := &http.Client{Timeout: 10 * time.Second}
	const maxRetries = 3
	var resp *http.Response

	for i := 0; i < maxRetries; i++ {
		resp, err = client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			return nil
		}

		time.Sleep(time.Duration((i+1)*2) * time.Second)
	}

	if err != nil {
		return fmt.Errorf("failed to send email after %d attempts, last error: %v", maxRetries, err)
	}
	defer resp.Body.Close()

	return fmt.Errorf("failed to send email: received status code %d", resp.StatusCode)
}
