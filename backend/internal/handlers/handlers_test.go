package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"

	"social-scribe/backend/internal/handlers"
	"social-scribe/backend/internal/middlewares"
	"social-scribe/backend/internal/models"
	"social-scribe/backend/internal/repositories"
	"social-scribe/backend/internal/scheduler"
	"social-scribe/backend/internal/services"
)

func init() {
	// Override GetUserByName to simulate "user does not exist" (successful signup)
	repositories.GetUserByName = func(name string) (*models.User, error) {
		return nil, nil
	}
	// Override InsertUser to return a fixed valid user ID.
	repositories.InsertUser = func(user models.User) (string, error) {
		return "507f1f77bcf86cd799439011", nil
	}
	// Override SetCache to simulate successful cache insertion.
	repositories.SetCache = func(key string, value interface{}, expiration time.Duration) error {
		return nil
	}
	repositories.GetCache = func(key string) (interface{}, bool) {
		return nil, false
	}

	// Override OTP generation and email sending.
	services.GenerateOTP = func() string {
		return "123456"
	}
	services.SendEmail = func(toEmail, message string) error {
		return nil
	}
	repositories.GetScheduledTasks = func() ([]models.ScheduledBlogData, error) {
		return []models.ScheduledBlogData{}, nil
	}
	repositories.StoreScheduledTask = func(task models.ScheduledBlogData) error {
		return nil
	}

	repositories.DeleteScheduledTask = func(task models.ScheduledBlogData) error {
		return nil
	}

	// Initialize scheduler properly for tests.
	taskScheduler := scheduler.NewScheduler()
	handlers.InitScheduler(taskScheduler)
}

func TestSignupUserHandler_Success(t *testing.T) {
	// Create a valid signup request.
	signupData := map[string]string{
		"UserName": "Test@Example.com", // Will be lowercased by handler.
		"PassWord": "password123",
	}
	body, _ := json.Marshal(signupData)
	req := httptest.NewRequest("POST", "/api/v1/user/signup", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	respRecorder := httptest.NewRecorder()
	handlers.SignupUserHandler(respRecorder, req)

	// Expect HTTP 201 Created.
	assert.Equal(t, http.StatusCreated, respRecorder.Code)
	assert.Equal(t, "application/json", respRecorder.Header().Get("Content-Type"))

	// Decode response JSON.
	var respUser models.User
	err := json.NewDecoder(respRecorder.Body).Decode(&respUser)
	assert.NoError(t, err)

	// Check that password is cleared in the response.
	assert.Equal(t, "", respUser.PassWord)
	// Email should be lowercased.
	assert.Equal(t, "test@example.com", respUser.UserName)
	// Check that a valid ObjectID was set.
	_, err = primitive.ObjectIDFromHex(respUser.Id.Hex())
	assert.NoError(t, err)
}

func TestSignupUserHandler_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest("POST", "/api/v1/user/signup", strings.NewReader("not a json"))
	req.Header.Set("Content-Type", "application/json")
	respRecorder := httptest.NewRecorder()

	handlers.SignupUserHandler(respRecorder, req)
	assert.Equal(t, http.StatusBadRequest, respRecorder.Code)
}

func TestSignupUserHandler_InvalidEmail(t *testing.T) {
	// Invalid email address.
	signupData := map[string]string{
		"UserName": "invalid-email",
		"PassWord": "password123",
	}
	body, _ := json.Marshal(signupData)
	req := httptest.NewRequest("POST", "/api/v1/user/signup", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	respRecorder := httptest.NewRecorder()

	handlers.SignupUserHandler(respRecorder, req)
	// Expect 400 Bad Request because email is invalid.
	assert.Equal(t, http.StatusBadRequest, respRecorder.Code)
}

func TestSignupUserHandler_ShortPassword(t *testing.T) {
	// Password too short.
	signupData := map[string]string{
		"UserName": "test@example.com",
		"PassWord": "short", // less than 8 chars.
	}
	body, _ := json.Marshal(signupData)
	req := httptest.NewRequest("POST", "/api/v1/user/signup", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	respRecorder := httptest.NewRecorder()

	handlers.SignupUserHandler(respRecorder, req)
	// Expect 400 Bad Request due to password length.
	assert.Equal(t, http.StatusBadRequest, respRecorder.Code)
}

// --- Test LoginUserHandler ---
// We'll override GetUserByName to simulate a valid user and password check.
// For simplicity, we assume a helper function repositories.HashPassword exists;
// if not, you can precompute a hash using bcrypt and return that.

func TestLoginUserHandler_Success(t *testing.T) {
	// Override GetUserByName for login.
	repositories.GetUserByName = func(name string) (*models.User, error) {
		// Simulate a user record with a hashed password.
		// For the purpose of this test, we assume the password is "password123".
		// In practice, you might use bcrypt.GenerateFromPassword to get a hash.
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		return &models.User{
			Id:       primitive.NewObjectID(),
			UserName: strings.ToLower(name),
			PassWord: string(hashedPassword),
		}, nil
	}

	// Prepare a valid login JSON body.
	loginData := map[string]string{
		"Username": "test@example.com",
		"Password": "password123",
	}
	body, _ := json.Marshal(loginData)
	req := httptest.NewRequest("POST", "/api/v1/user/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	respRecorder := httptest.NewRecorder()

	handlers.LoginUserHandler(respRecorder, req)
	// Expect HTTP 202 Accepted on successful login.
	assert.Equal(t, http.StatusAccepted, respRecorder.Code)
	assert.Equal(t, "application/json", respRecorder.Header().Get("Content-Type"))

	// Optionally decode response and check that password is cleared.
	var respUser models.User
	err := json.NewDecoder(respRecorder.Body).Decode(&respUser)
	assert.NoError(t, err)
	assert.Equal(t, "", respUser.PassWord)
}

func TestLoginUserHandler_EmptyBody(t *testing.T) {
	req := httptest.NewRequest("POST", "/api/v1/user/login", nil)
	respRecorder := httptest.NewRecorder()

	handlers.LoginUserHandler(respRecorder, req)
	assert.Equal(t, http.StatusBadRequest, respRecorder.Code)
}

// --- Test GetUserInfoHandler ---
// We override GetUserById to simulate fetching a user.

func TestGetUserInfoHandler_Success(t *testing.T) {
	// Override GetUserById to return a dummy user.
	repositories.GetUserById = func(userId string) (*models.User, error) {
		return &models.User{
			Id:       primitive.NewObjectID(),
			UserName: "test@example.com",
			PassWord: "hashedpass", // This will be cleared in response.
		}, nil
	}

	req := httptest.NewRequest("GET", "/api/v1/user/getinfo", nil)
	// Simulate that authentication middleware injected a user ID into context.
	dummyID := primitive.NewObjectID().Hex()
	ctx := context.WithValue(req.Context(), middlewares.UserIDKey, dummyID)
	req = req.WithContext(ctx)

	respRecorder := httptest.NewRecorder()
	handlers.GetUserInfoHandler(respRecorder, req)

	assert.Equal(t, http.StatusOK, respRecorder.Code)
	// Optionally decode response to verify user data.
	var userResp models.User
	err := json.NewDecoder(respRecorder.Body).Decode(&userResp)
	assert.NoError(t, err)
	assert.Equal(t, "test@example.com", userResp.UserName)
}

func TestGetUserInfoHandler_UserNotFound(t *testing.T) {
	// Override GetUserById to return nil.
	repositories.GetUserById = func(userId string) (*models.User, error) {
		return nil, nil
	}
	req := httptest.NewRequest("GET", "/api/v1/user/getinfo", nil)
	dummyID := primitive.NewObjectID().Hex()
	ctx := context.WithValue(req.Context(), middlewares.UserIDKey, dummyID)
	req = req.WithContext(ctx)
	respRecorder := httptest.NewRecorder()

	handlers.GetUserInfoHandler(respRecorder, req)
	assert.Equal(t, http.StatusNotFound, respRecorder.Code)
}
