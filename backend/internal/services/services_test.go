package services

import (
	"fmt"
	"os"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestInvokeAi_Success(t *testing.T) {
	os.Setenv("GEMINI_API_KEY", "dummy_key")
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	mockResponse := `{"candidates": [{"content": {"parts": [{"text": "This is a test AI response"}]}}]}`
	expectedURL := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent?key=%s", "dummy_key")

	httpmock.RegisterResponder("POST", expectedURL, httpmock.NewStringResponder(200, mockResponse))

	response, err := invokeAi("Test prompt")
	assert.NoError(t, err)
	assert.Equal(t, "This is a test AI response", response)
}

func TestInvokeAi_APIError(t *testing.T) {
	os.Setenv("GEMINI_API_KEY", "dummy_key")
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	expectedURL := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent?key=%s", "dummy_key")
	httpmock.RegisterResponder("POST", expectedURL, httpmock.NewStringResponder(500, "Internal Server Error"))

	response, err := invokeAi("Test prompt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error")
	assert.Empty(t, response)
}

func TestInvokeAi_InvalidJSONResponse(t *testing.T) {
	os.Setenv("GEMINI_API_KEY", "dummy_key")
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	expectedURL := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent?key=%s", "dummy_key")
	httpmock.RegisterResponder("POST", expectedURL, httpmock.NewStringResponder(200, "not a json"))

	response, err := invokeAi("Test prompt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal")
	assert.Empty(t, response)
}

func TestGetUserURN_Success(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	mockResponse := `{"sub": "12345"}`
	httpmock.RegisterResponder("GET", "https://api.linkedin.com/v2/userinfo", httpmock.NewStringResponder(200, mockResponse))

	response, err := getUserURN("dummy_access_token")
	assert.NoError(t, err)
	assert.Equal(t, "urn:li:person:12345", response)
}

func TestGetUserURN_InvalidResponse(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "https://api.linkedin.com/v2/userinfo", httpmock.NewStringResponder(200, "invalid json"))

	response, err := getUserURN("dummy_access_token")
	assert.Error(t, err)
	assert.Empty(t, response)
}

func TestLinkedPostHandler_Success(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "https://api.linkedin.com/v2/userinfo", httpmock.NewStringResponder(200, `{"sub": "12345"}`))
	httpmock.RegisterResponder("POST", "https://api.linkedin.com/v2/ugcPosts", httpmock.NewStringResponder(201, ""))

	err := linkedPostHandler("Test LinkedIn post", "dummy_access_token")
	assert.NoError(t, err)
}

func TestGenerateOTP(t *testing.T) {
	otp := GenerateOTP()
	assert.Len(t, otp, 6)
}
