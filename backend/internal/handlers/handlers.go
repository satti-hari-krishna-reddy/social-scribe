package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/mail"
	"os"

	"social-scribe/backend/internal/middlewares"
	"social-scribe/backend/internal/models"
	repo "social-scribe/backend/internal/repositories"
	"social-scribe/backend/internal/scheduler"
	"social-scribe/backend/internal/services"
	"social-scribe/backend/internal/utils"

	"strings"
	"time"

	"github.com/dghubble/oauth1"
	"github.com/dghubble/oauth1/twitter"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/linkedin"
)

var twitterConfig = &oauth1.Config{}
var linkedinConfig = &oauth2.Config{}
var frontendURL = os.Getenv("FRONTEND_URL")
var taskScheduler *scheduler.Scheduler

func init() {
	if os.Getenv("TEST_ENV") == "true" {
		return // Skip loading .env in tests ?? Hmmm, is there a beter way ?
	}
	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("[INFO] No .env file found, relying on system environment variables")
	}
	frontendURL = os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		log.Println("[WARN] FRONTEND_URL not set, using default")
		frontendURL = "http://localhost:5173"
	}

	twitterConfig = &oauth1.Config{
		ConsumerKey:    os.Getenv("TWITTER_CONSUMER_KEY"),
		ConsumerSecret: os.Getenv("TWITTER_CONSUMER_SECRET"),
		CallbackURL:    os.Getenv("TWITTER_CALLBACK_URL"),
		Endpoint:       twitter.AuthorizeEndpoint,
	}
	linkedinConfig = &oauth2.Config{
		ClientID:     os.Getenv("LINKEDIN_CLIENT_ID"),
		ClientSecret: os.Getenv("LINKEDIN_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("LINKEDIN_CALLBACK_URL"),
		Scopes:       []string{"openid", "profile", "email", "w_member_social"},
		Endpoint:     linkedin.Endpoint,
	}

	services.InitTwitterConfig(twitterConfig)

}

func InitScheduler(s *scheduler.Scheduler) {
	taskScheduler = s
}

func SignupUserHandler(resp http.ResponseWriter, req *http.Request) {
	if req.Body == nil {
		http.Error(resp, `{"error": "Request body is empty"}`, http.StatusBadRequest)
		return
	}

	var user models.User
	if err := json.NewDecoder(req.Body).Decode(&user); err != nil {
		http.Error(resp, `{"error": "Unable to decode JSON"}`, http.StatusBadRequest)
		return
	}

	user.UserName = strings.TrimSpace(strings.ToLower(user.UserName))
	user.PassWord = strings.TrimSpace(user.PassWord)

	// Validate email using net/mail.
	if _, err := mail.ParseAddress(user.UserName); err != nil {
		http.Error(resp, `{"error": "Invalid email address"}`, http.StatusBadRequest)
		return
	}

	if len(user.PassWord) < 8 || len(user.PassWord) > 128 {
		http.Error(resp, `{"error": "Password must be between 8 and 128 characters"}`, http.StatusBadRequest)
		return
	}

	// Check if user already exists.
	existingUser, err := repo.GetUserByName(user.UserName)
	if err != nil {
		http.Error(resp, `{"error": "Internal server error"}`, http.StatusInternalServerError)
		log.Printf("[ERROR] Checking existing user: %v", err)
		return
	}
	if existingUser != nil {
		http.Error(resp, `{"message": "Username already taken"}`, http.StatusConflict)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.PassWord), bcrypt.DefaultCost)
	if err != nil {
		http.Error(resp, `{"error": "Internal server error"}`, http.StatusInternalServerError)
		log.Printf("[ERROR] Hashing password for user '%s': %v", user.UserName, err)
		return
	}

	user.Verified = false
	user.LinkedinVerified = false
	user.EmailVerified = false
	user.HashnodeVerified = false
	user.XVerified = false
	user.PassWord = string(hashedPassword)

	userId, err := repo.InsertUser(user)
	if err != nil {
		log.Printf("[ERROR] Creating user %s: %v", user.UserName, err)
		http.Error(resp, `{"error": "Failed to create user"}`, http.StatusInternalServerError)
		return
	}
	objectId, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		http.Error(resp, `{"error": "Invalid user ID format"}`, http.StatusInternalServerError)
		return
	}
	user.Id = objectId

	sessionToken := uuid.New().String()
	expiration := time.Now().Add(24 * time.Hour)
	if err := repo.SetCache(sessionToken, user.Id, 24*time.Hour); err != nil {
		http.Error(resp, `{"error": "Failed to create session"}`, http.StatusInternalServerError)
		return
	}

	http.SetCookie(resp, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		HttpOnly: true,
		Path:     "/",
		Secure:   false,
		Expires:  expiration,
	})
	cacheKey := fmt.Sprintf("email_otp_%s", userId)
	otp := services.GenerateOTP()
	err = repo.SetCache(cacheKey, otp, 24*time.Hour)
	if err != nil {
		log.Printf("[ERROR] Failed to store OTP in cache for user %s: %v", user.UserName, err)
		http.Error(resp, `{"error": "Failed to store OTP"}`, http.StatusInternalServerError)
		return
	}
	message := fmt.Sprintf("Your OTP is: %s \n Valid for next 24 hours", otp)
	// this is just a temporary work around to use the existing heap based queue system
	// to send emails asynchronously and we need a better way to do this in the future
	emailTask := models.ScheduledBlogData{}
	shareTime := models.ScheduledBlog{}
	shareTime.ScheduledTime = time.Now()
	emailTask.EmailId = user.UserName
	emailTask.Message = message
	emailTask.ScheduledBlog = shareTime
	emailTask.UserID = userId

	err = taskScheduler.AddTask(emailTask)
	if err != nil {
		log.Printf("[ERROR] Failed adding email task to the queue, reason: %s", err)
		resp.WriteHeader(http.StatusInternalServerError)
		resp.Write([]byte(`{"success" : false}`))
	}

	user.PassWord = ""
	responseJson, err := json.Marshal(user)
	if err != nil {
		http.Error(resp, `{"error": "Failed to process user data"}`, http.StatusInternalServerError)
		return
	}
	// Use CSRF helper to get or create a token.
	csrfToken, err := utils.GetOrCreateCsrfToken(userId)
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		resp.Write([]byte(`{"success": false, "reason": "Failed to create CSRF token"}`))
		return
	}

	resp.Header().Set("X-Csrf-Token", csrfToken)

	log.Printf("[INFO] User '%s' registered with ID: %s", user.UserName, userId)
	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(http.StatusCreated)
	resp.Write(responseJson)
}

func LoginUserHandler(resp http.ResponseWriter, req *http.Request) {
	if req.Body == nil {
		http.Error(resp, `{"error": "Failed to parse login credentials: body is empty"}`, http.StatusBadRequest)
		return
	}
	data := models.LoginStruct{}

	err := json.NewDecoder(req.Body).Decode(&data)
	if err != nil {
		http.Error(resp, `{"error": "Bad request: unable to decode JSON"}`, http.StatusBadRequest)
		return
	}

	data.Username = strings.ToLower(strings.TrimSpace(data.Username))
	if len(data.Username) < 4 || len(data.Username) > 64 {
		http.Error(resp, `{"error": "Username is should in range of minimum 4 to maximum 64 characters}`, http.StatusBadGateway)
	}
	if len(data.Password) > 128 {
		http.Error(resp, `{"error" : "password is too long, the maximum allowed length is 128 chars"}`, http.StatusBadGateway)
	}
	user, err := repo.GetUserByName(data.Username)
	if user == nil {
		http.Error(resp, `{"success": false, "reason": "Username and/or password is incorrect"}`, http.StatusBadRequest)
		return
	}
	if err != nil {
		log.Printf("[ERROR] Failed to get user for the username %s and the error is %s", data.Username, err)
		http.Error(resp, `{"error" : "Internal server error"}`, http.StatusInternalServerError)
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.PassWord), []byte(data.Password))
	if err != nil {
		http.Error(resp, `{"success": false, "reason": "Username and/or password is incorrect"}`, http.StatusBadRequest)
		return
	}

	sessionToken := uuid.New().String()
	expiration := time.Now().Add(24 * time.Hour)
	err = repo.SetCache(sessionToken, user.Id, 24*time.Hour)
	if err != nil {
		http.Error(resp, `{"error": "Failed to create session"}`, http.StatusInternalServerError)
		return
	}

	http.SetCookie(resp, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		HttpOnly: true,
		Path:     "/",
		Secure:   false,
		Expires:  expiration,
	})

	user.PassWord = ""
	responseJson, err := json.Marshal(user)
	if err != nil {
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Failed unpacking user"}`))
		return
	}
	// Use CSRF helper to get or create a token.
	csrfToken, err := utils.GetOrCreateCsrfToken(user.Id.Hex())
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		resp.Write([]byte(`{"success": false, "reason": "Failed to create CSRF token"}`))
		return
	}

	resp.Header().Set("X-Csrf-Token", csrfToken)

	resp.WriteHeader(http.StatusAccepted)
	resp.Header().Set("Content-Type", "application/json")
	resp.Write([]byte(responseJson))
}

func LogoutUserHandler(resp http.ResponseWriter, req *http.Request) {
	userId, ok := req.Context().Value(middlewares.UserIDKey).(string)
	if !ok {
		http.Error(resp, "Unauthorized: User ID not found", http.StatusUnauthorized)
		return
	}

	err := repo.DeleteCache(userId)
	if err != nil {
		log.Printf("[ERROR] Failed to delete session for user %s: %v", userId, err)
		http.Error(resp, "Internal server error", http.StatusInternalServerError)
		return
	}

	cacheKey := fmt.Sprintf("CSRF_%s", userId)
	err = repo.DeleteCache(cacheKey)
	if err != nil {
		log.Printf("[ERROR] Failed to delete CSRF token for user %s: %s", userId, err)
	}

	http.SetCookie(resp, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	log.Printf("[INFO] User with ID %s logged out successfully", userId)
	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(http.StatusOK)
	resp.Write([]byte(`{"success": true}`))
}

func GetUserInfoHandler(resp http.ResponseWriter, req *http.Request) {
	userId, ok := req.Context().Value(middlewares.UserIDKey).(string)
	if !ok {
		http.Error(resp, "Unauthorized: User ID not found", http.StatusUnauthorized)
		return
	}

	user, err := repo.GetUserById(userId)
	if err != nil {
		log.Printf("[ERROR] Failed to find user for the id: %s and error is %s", userId, err)
		http.Error(resp, `{"error": ""}`, http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(resp, `{"error": "user id is not valid"}`, http.StatusNotFound)
		return
	}
	user.PassWord = ""
	user.HashnodePAT = ""
	user.LinkedInOauthKey = ""
	user.XOAuthToken = ""
	user.XOAuthSecret = ""

	responseJson, err := json.Marshal(user)
	if err != nil {
		resp.WriteHeader(http.StatusUnauthorized)
		resp.Write([]byte(`{"success": false, "reason": "Failed unpacking"}`))
		return
	}

	// Use CSRF helper to get or create a token.
	csrfToken, err := utils.GetOrCreateCsrfToken(userId)
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		resp.Write([]byte(`{"success": false, "reason": "Failed to create CSRF token"}`))
		return
	}

	resp.Header().Set("X-Csrf-Token", csrfToken)
	resp.WriteHeader(http.StatusOK)
	resp.Write(responseJson)
}

func GetUserNotificationsHandler(resp http.ResponseWriter, req *http.Request) {
	userId, ok := req.Context().Value(middlewares.UserIDKey).(string)
	if !ok {
		http.Error(resp, "Unauthorized: User ID not found", http.StatusUnauthorized)
		return
	}
	user, err := repo.GetUserById(userId)
	if err != nil {
		log.Printf("[ERROR] Failed to find user for the id: %s and error is %s", userId, err)
		http.Error(resp, `{"error": ""}`, http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(resp, `{"error": "user id is not valid"}`, http.StatusNotFound)
		return
	}
	respone := map[string]interface{}{
		"notifications": user.Notifications,
	}
	responseJson, err := json.Marshal(respone)
	if err != nil {
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Failed unpacking"}`))
		return
	}

	resp.WriteHeader(200)
	resp.Write(responseJson)

}

func ClearUserNotificationsHandler(resp http.ResponseWriter, req *http.Request) {
	userId, ok := req.Context().Value(middlewares.UserIDKey).(string)
	if !ok {
		http.Error(resp, "Unauthorized: User ID not found", http.StatusUnauthorized)
		return
	}
	user, err := repo.GetUserById(userId)
	if err != nil {
		log.Printf("[ERROR] failed to get user for the id: %s and the error is %s", userId, err)
		resp.WriteHeader(500)
		resp.Write([]byte(`{"success" : false}`))
		return
	}
	if user == nil {
		resp.WriteHeader(401)
		resp.Write([]byte(`"success" : false, "reason" : "invalid user id"`))
		return

	}
	user.Notifications = []string{}
	err = repo.UpdateUser(userId, user)
	if err != nil {
		log.Printf("[ERROR] failed to update user with id: %s", userId)
		resp.WriteHeader(500)
		resp.Write([]byte(`{"success" : false, "reason" : "}`))
		return
	}
	resp.WriteHeader(200)
	resp.Write([]byte(`{"success" : true, "message" : "notifications cleared sucessfully"}`))
}

func GetUserBlogsHandler(resp http.ResponseWriter, req *http.Request) {
	userId, ok := req.Context().Value(middlewares.UserIDKey).(string)
	if !ok {
		http.Error(resp, "Unauthorized: User ID not found", http.StatusUnauthorized)
		return
	}
	user, err := repo.GetUserById(userId)
	if err != nil {
		log.Printf("[ERROR] Failed to get user for id: %s - %v", userId, err)
		http.Error(resp, "Internal server error", http.StatusInternalServerError)
		return
	}
	if user == nil {
		log.Printf("[ERROR] User with id: %s not found", userId)
		http.Error(resp, "User not found", http.StatusNotFound)
		return
	}

	category := strings.ToLower(req.URL.Query().Get("category"))
	if category == "" {
		category = "all"
	} else if category != "all" && category != "scheduled" && category != "shared" {
		http.Error(resp, "Invalid category", http.StatusBadRequest)
		return
	}

	var responseBytes []byte
	var jsonErr error

	switch category {
	case "scheduled":
		responseBytes, jsonErr = json.Marshal(user.ScheduledBlogs)
	case "shared":
		responseBytes, jsonErr = json.Marshal(user.SharedBlogs)
	default:
		// Handle "all" case with GraphQL
		endpoint := "https://gql.hashnode.com"
		query := models.GraphQLQuery{
			Query: fmt.Sprintf(`
                query Publication {
                    publication(host: "%s") {
                        posts(first: 0) {
                            edges {
                                node {
                                    title
                                    url
                                    id
                                    coverImage { url }
                                    author { name }
                                    readTimeInMinutes
                                }
                            }
                        }
                    }
                }`, user.HashnodeBlog),
		}

		queryBytes, err := json.Marshal(query)
		if err != nil {
			log.Printf("[ERROR] Failed to marshal query: %v", err)
			http.Error(resp, "Internal server error", http.StatusInternalServerError)
			return
		}

		headers := map[string]string{"Content-Type": "application/json"}
		gqlResponse, err := services.MakePostRequest(endpoint, queryBytes, headers)
		if err != nil {
			log.Printf("[ERROR] Failed to make request: %v", err)
			http.Error(resp, "Internal server error", http.StatusInternalServerError)
			return
		}

		var gqlData models.GraphQLResponse
		if err := json.Unmarshal(gqlResponse, &gqlData); err != nil {
			log.Printf("[ERROR] Failed to unmarshal response: %v", err)
			http.Error(resp, "Internal server error", http.StatusInternalServerError)
			return
		}

		var posts []models.PostNode
		for _, edge := range gqlData.Data.Publication.Posts.Edges {
			posts = append(posts, edge.Node)
		}
		responseBytes, jsonErr = json.Marshal(posts)
	}

	// Handle JSON marshaling errors
	if jsonErr != nil {
		log.Printf("[ERROR] Failed to marshal response: %v", jsonErr)
		http.Error(resp, "Internal server error", http.StatusInternalServerError)
		return
	}

	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(http.StatusOK)
	resp.Write([]byte(fmt.Sprintf(`{"success": true, "blogs": %s}`, string(responseBytes))))
}

func ConnectXhandler(resp http.ResponseWriter, req *http.Request) {
	userId, ok := req.Context().Value(middlewares.UserIDKey).(string)
	if !ok {
		http.Error(resp, "Unauthorized: User ID not found", http.StatusUnauthorized)
		return
	}
	user, err := repo.GetUserById(userId)
	if err != nil {
		log.Printf("[ERROR] Failed to get user for the id: %s and error is %s", userId, err)
		return
	}
	if user == nil {
		log.Printf("[ERROR] User with id: %s not found", userId)
		return
	}

	requestToken, requestSecret, err := twitterConfig.RequestToken()
	if err != nil {
		fmt.Printf("error: %v", err)
		http.Error(resp, "Failed to get request token", http.StatusInternalServerError)
		return
	}
	user.XOAuthToken = requestToken
	user.XOAuthSecret = requestSecret
	err = repo.UpdateUser(userId, user)
	if err != nil {
		log.Printf("[ERROR] Failed to update user with id: %s and error is %s", userId, err)
		return
	}

	authorizationURL, err := twitterConfig.AuthorizationURL(requestToken)
	if err != nil {
		http.Error(resp, "Failed to get authorization URL", http.StatusInternalServerError)
		return
	}

	http.Redirect(resp, req, authorizationURL.String(), http.StatusFound)
}

func XcallbackHandler(resp http.ResponseWriter, req *http.Request) {
	userID, ok := req.Context().Value(middlewares.UserIDKey).(string)
	if !ok {
		http.Error(resp, "Unauthorized: User ID not found", http.StatusUnauthorized)
		return
	}
	user, err := repo.GetUserById(userID)
	if err != nil {
		http.Error(resp, "Failed to get user", http.StatusInternalServerError)
		log.Printf("[ERROR] Failed to get user for the id: %s and error is %s", userID, err)

		return
	}
	if user == nil {
		http.Error(resp, "User not found", http.StatusNotFound)
		log.Printf("[ERROR] User with id: %s not found", userID)
		return
	}

	token := user.XOAuthToken
	secret := user.XOAuthSecret

	requestTokenData := &oauth1.Token{Token: token, TokenSecret: secret}
	verifier := req.URL.Query().Get("oauth_verifier")
	if verifier == "" {
		log.Printf("[ERROR] Missing OAuth verifier for user with id: %s", userID)
		http.Error(resp, "Missing OAuth verifier", http.StatusBadRequest)
		return
	}
	accessToken, accessSecret, err := twitterConfig.AccessToken(requestTokenData.Token, requestTokenData.TokenSecret, verifier)
	if err != nil {
		log.Printf("[ERROR] Failed to get access token for user with id: %s and error is %s", userID, err)
		http.Error(resp, "Failed to get access token", http.StatusInternalServerError)
		return
	}
	user.XOAuthToken = accessToken
	user.XOAuthSecret = accessSecret
	user.XVerified = true
	if (user.XVerified || user.LinkedinVerified) && user.HashnodeVerified {
		user.Verified = true
	} else {
		user.Verified = false
	}
	err = repo.UpdateUser(userID, user)
	if err != nil {
		http.Error(resp, "Failed to update user", http.StatusInternalServerError)
		log.Printf("[ERROR] Failed to update user with id: %s and error is %s", userID, err)
		return
	}

	log.Printf("[INFO] User with ID %s connected to X(twitter) Successfully", user.Id)
	redirectUrl := fmt.Sprintf("%s/verification", frontendURL)
	http.Redirect(resp, req, redirectUrl, http.StatusSeeOther)
}

func ConnectLinkedInHandler(resp http.ResponseWriter, req *http.Request) {
	userId, ok := req.Context().Value(middlewares.UserIDKey).(string)
	if !ok {
		http.Error(resp, "Unauthorized: User ID not found", http.StatusUnauthorized)
		return
	}
	user, err := repo.GetUserById(userId)
	if err != nil {
		log.Printf("[ERROR] Failed to get user for the id: %s and error is %s", userId, err)
		return
	}
	if user == nil {
		log.Printf("[ERROR] User with id: %s not found", userId)
		return
	}
	state := uuid.New().String()
	err = repo.SetCache(state, userId, 10*time.Minute)
	if err != nil {
		log.Printf("[ERROR] Failed to store state in cache: %v", err)
		http.Error(resp, "Failed to store state in cache", http.StatusInternalServerError)
		return
	}
	expiration := time.Now().Add(10 * time.Minute)

	http.SetCookie(resp, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		HttpOnly: true,
		Path:     "/",
		Secure:   false,
		Expires:  expiration,
	})

	authURL := linkedinConfig.AuthCodeURL(state)
	http.Redirect(resp, req, authURL, http.StatusFound)
}

func LinkedCallbackHandler(resp http.ResponseWriter, req *http.Request) {
	queryState := req.URL.Query().Get("state")
	stateCookie, err := req.Cookie("oauth_state")
	log.Printf("State cookie: %v", stateCookie)

	if stateCookie == nil {
		log.Printf("[ERROR] Missing state cookie")
	}
	if err != nil || stateCookie.Value != queryState {
		log.Printf("[ERROR] Invalid state parameter")
		http.Error(resp, "Invalid state parameter", http.StatusForbidden)
		return
	}
	cacheItem, exists := repo.GetCache(stateCookie.Value)
	if !exists {
		log.Printf("[ERROR] Invalid state parameter")
		http.Error(resp, "Invalid state parameter", http.StatusForbidden)
		return
	}

	userId, ok := cacheItem.(models.CacheItem)
	if !ok {
		log.Printf("[ERROR] Failed to cast cached item to CacheItem")
		http.Error(resp, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	userIdStr, ok := userId.Value.(string)
	if !ok {
		log.Printf("[ERROR] Failed to cast CacheItem.Value to string")
		http.Error(resp, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = repo.DeleteCache(stateCookie.Value)
	if err != nil {
		log.Printf("[WARN] Failed to delete state from cache for the user id: %s and error is %s", userId, err)
	}

	user, err := repo.GetUserById(userIdStr)
	if err != nil {
		log.Printf("[ERROR] Failed to get user for the id: %s and error is %s", userId, err)
		http.Error(resp, "User not found", http.StatusNotFound)
		return
	}
	if user == nil {
		log.Printf("[ERROR] User with id: %s not found", userId)
		http.Error(resp, "User not found", http.StatusNotFound)
		return
	}

	code := req.URL.Query().Get("code")
	if code == "" {
		log.Printf("[ERROR] Missing authorization code")
		http.Error(resp, "Missing authorization code", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	token, err := linkedinConfig.Exchange(ctx, code)
	if err != nil {
		http.Error(resp, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}
	user.LinkedInOauthKey = token.AccessToken
	user.LinkedinVerified = true
	if (user.XVerified || user.LinkedinVerified) && user.HashnodeVerified {
		user.Verified = true
	} else {
		user.Verified = false
	}
	err = repo.UpdateUser(userIdStr, user)
	if err != nil {
		log.Printf("[ERROR] Failed to update user with id: %s and error is %s", userId, err)
		http.Error(resp, "Failed to update user", http.StatusInternalServerError)
		return
	}
	log.Printf("[INFO] User with ID %s connected to LinkedIn Successfully", userIdStr)

	// Redirect the user back to the frontend
	redirectUrl := fmt.Sprintf("%s/verification", frontendURL)
	http.Redirect(resp, req, redirectUrl, http.StatusSeeOther)
}

func ValidateLogin(req *http.Request) (string, error) {
	cookie, err := req.Cookie("session_token")
	if err != nil {
		return "", fmt.Errorf("missing session token")
	}

	sessionData, exists := repo.GetCache(cookie.Value)
	if !exists {
		return "", fmt.Errorf("invalid or expired session")
	}

	session, ok := sessionData.(models.CacheItem)
	if !ok {
		return "", fmt.Errorf("invalid session data format")
	}

	if session.ExpiresAt.Before(time.Now()) {
		return "", fmt.Errorf("session expired")
	}

	// session.Value is actually a primitive.ObjectID, convert it to string.
	oid, ok := session.Value.(primitive.ObjectID)
	if !ok {
		return "", fmt.Errorf("invalid session user id format")
	}
	return oid.Hex(), nil
}

func VerifyHashnodeHandler(w http.ResponseWriter, r *http.Request) {
	endpoint := "https://gql.hashnode.com"
	userId, ok := r.Context().Value(middlewares.UserIDKey).(string)
	if !ok {
		http.Error(w, "Unauthorized: User ID not found", http.StatusUnauthorized)
		return
	}
	user, err := repo.GetUserById(userId)
	if err != nil {
		log.Printf("[ERROR] Failed to get user for the id: %s and error is %s", userId, err)
		return
	}
	if user == nil {
		log.Printf("[ERROR] User with id: %s not found", userId)
		return
	}

	var hashnodeKey models.HashnodeKey
	err = json.NewDecoder(r.Body).Decode(&hashnodeKey)
	if err != nil {
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		return
	}
	if hashnodeKey.Key == "" {
		http.Error(w, "Missing Hashnode API key", http.StatusBadRequest)
		return
	}

	query := `{"query":"query Me { me { publications(first:1) { edges { node { url id } } } } }"}`

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer([]byte(query)))
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", hashnodeKey.Key)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "Failed to make request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Invalid Hashnode API key", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read response", http.StatusInternalServerError)
		return
	}

	var response struct {
		Data struct {
			Me struct {
				Publications struct {
					Edges []struct {
						Node struct {
							URL string `json:"url"`
							ID  string `json:"id"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"publications"`
			} `json:"me"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		http.Error(w, "Failed to parse response JSON", http.StatusInternalServerError)
		return
	}

	// Check if we have at least one publication
	if len(response.Data.Me.Publications.Edges) == 0 {
		http.Error(w, "No publications found", http.StatusNotFound)
		return
	}

	// Extract `url` and `id`
	node := response.Data.Me.Publications.Edges[0].Node
	url := strings.ReplaceAll(node.URL, "https://", "")
	id := node.ID

	user.HashnodePAT = hashnodeKey.Key
	user.HashnodeVerified = true
	user.HashnodeBlog = url
	if (user.XVerified || user.LinkedinVerified) && user.HashnodeVerified {
		user.Verified = true
	} else {
		user.Verified = false
	}
	err = repo.UpdateUser(userId, user)
	if err != nil {
		log.Printf("[ERROR] Failed to update user with id: %s and error is %s", userId, err)
		return
	}
	fmt.Printf(`{"success": true, "url": "%s", "id": "%s"}`, url, id)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success": true}`))
}

func ShareBlogHandler(resp http.ResponseWriter, req *http.Request) {
	userId, ok := req.Context().Value(middlewares.UserIDKey).(string)
	if !ok {
		http.Error(resp, "Unauthorized: User ID not found", http.StatusUnauthorized)
		return
	}
	user, err := repo.GetUserById(userId)
	if err != nil {
		log.Printf("[ERROR] Failed to get user for the id: %s and error is %s", userId, err)
		return
	}
	if user == nil {
		log.Printf("[ERROR] User with id: %s not found", userId)
		return
	}
	if !user.Verified {
		http.Error(resp, "User is not verified", http.StatusForbidden)
		return
	}

	var requestBody struct {
		Id        string   `json:"id"`
		Platforms []string `json:"platforms"`
	}
	if err := json.NewDecoder(req.Body).Decode(&requestBody); err != nil {
		http.Error(resp, "Invalid request body", http.StatusBadRequest)
		return
	}

	blogId := requestBody.Id
	if len(blogId) == 0 {
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "missing blog id in the request"}`))
		return
	}

	err = services.ProcessSharedBlog(user, blogId, requestBody.Platforms)
	if err != nil {
		log.Printf("[ERROR] Failed to share blog: %v", err)
		http.Error(resp, "Failed to share blog", http.StatusInternalServerError)
		return
	}
	log.Printf("[INFO] Blog with ID %s shared successfully by user with ID %s", blogId, userId)

	resp.WriteHeader(http.StatusOK)
	resp.Write([]byte(`{"success": true}`))
}

func ScheduleBlogHandler(resp http.ResponseWriter, req *http.Request) {
	userId, ok := req.Context().Value(middlewares.UserIDKey).(string)
	if !ok {
		http.Error(resp, "Unauthorized: User ID not found", http.StatusUnauthorized)
		return
	}
	user, err := repo.GetUserById(userId)
	if err != nil {
		log.Printf("[ERROR] Failed to get user for the id: %s and error is %s", userId, err)
		http.Error(resp, "Internal server error", http.StatusInternalServerError)
		return
	}
	if user == nil {
		log.Printf("[ERROR] User with id: %s not found", userId)
		http.Error(resp, "User not found", http.StatusNotFound)
		return
	}
	if !user.Verified {
		log.Printf("[ERROR] User with id: %s is not verified", userId)
		http.Error(resp, "User is not verified", http.StatusForbidden)
		return
	}
	var blogData models.ScheduledBlogData
	err = json.NewDecoder(req.Body).Decode(&blogData)
	if err != nil {
		http.Error(resp, "Failed to parse JSON", http.StatusBadRequest)
		return
	}
	blogData.UserID = userId
	err = blogData.ScheduledBlog.Validate()
	if err != nil {
		http.Error(resp, err.Error(), http.StatusBadRequest)
		return
	}
	//check if the user has already scheduled the blog
	for i := range user.ScheduledBlogs {
		if user.ScheduledBlogs[i].Id == blogData.ScheduledBlog.Id {
			http.Error(resp, "Blog already scheduled", http.StatusBadRequest)
			return
		}
	}

	err = taskScheduler.AddTask(blogData)
	if err != nil {
		http.Error(resp, "Failed to store scheduled task", http.StatusInternalServerError)
		return
	}

	user.ScheduledBlogs = append(user.ScheduledBlogs, blogData.ScheduledBlog)
	err = repo.UpdateUser(userId, user)
	if err != nil {
		log.Printf("[ERROR] Failed to update user with id: %s and error is %s", userId, err)
		http.Error(resp, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("[INFO] Blog with ID %s scheduled successfully by user with ID %s", blogData.ScheduledBlog.Id, userId)
	resp.WriteHeader(http.StatusOK)
	resp.Write([]byte(`{"success": true}`))

}

func CancelScheduledBlogHandler(resp http.ResponseWriter, req *http.Request) {
	userId, ok := req.Context().Value(middlewares.UserIDKey).(string)
	if !ok {
		http.Error(resp, "Unauthorized: User ID not found", http.StatusUnauthorized)
		return
	}
	user, err := repo.GetUserById(userId)
	if err != nil {
		log.Printf("[ERROR] Failed to get user for the id: %s and error is %s", userId, err)
		http.Error(resp, "Internal server error", http.StatusInternalServerError)
		return
	}
	if user == nil {
		log.Printf("[ERROR] User with id: %s not found", userId)
		http.Error(resp, "User not found", http.StatusNotFound)
		return
	}
	if !user.Verified {
		log.Printf("[ERROR] User with id: %s is not verified", userId)
		http.Error(resp, "User is not verified", http.StatusForbidden)
		return
	}
	var requestBody struct {
		Id string `json:"id"`
	}
	if err := json.NewDecoder(req.Body).Decode(&requestBody); err != nil {
		http.Error(resp, "Invalid request body", http.StatusBadRequest)
		return
	}
	blogId := requestBody.Id
	if len(blogId) == 0 {
		http.Error(resp, "Missing blog id", http.StatusBadRequest)
		return
	}
	var updatedScheduledBlogs []models.ScheduledBlog
	for _, blog := range user.ScheduledBlogs {
		if blog.Id == blogId {
			continue
		}
		updatedScheduledBlogs = append(updatedScheduledBlogs, blog)
	}
	user.ScheduledBlogs = updatedScheduledBlogs
	err = repo.UpdateUser(userId, user)
	if err != nil {

		log.Printf("[ERROR] Failed to update user with id: %s and error is %s", userId, err)
		http.Error(resp, "Internal server error", http.StatusInternalServerError)
		return
	}
	err = taskScheduler.RemoveTask(blogId)
	if err != nil {
		log.Printf("[ERROR] Failed to remove scheduled task with id: %s and error is %s", blogId, err)
		http.Error(resp, "Internal server error", http.StatusInternalServerError)
		return
	}
	log.Printf("[INFO] Scheduled blog with ID %s cancelled successfully by user with ID %s", blogId, userId)
	resp.WriteHeader(http.StatusOK)
	resp.Write([]byte(`{"success": true}`))
}

func VerifyEmailHandler(resp http.ResponseWriter, req *http.Request) {
	if req.Body == nil {
		resp.Header().Set("Content-Type", "application/json")
		resp.WriteHeader(http.StatusBadRequest)
		resp.Write([]byte(`{"success": false, "message": "Request body is empty"}`))
		return
	}

	userId, err := ValidateLogin(req)
	if err != nil {
		resp.Header().Set("Content-Type", "application/json")
		resp.WriteHeader(http.StatusUnauthorized)
		resp.Write([]byte(`{"success": false, "message": "Unauthorized"}`))
		return
	}

	user, err := repo.GetUserById(userId)
	if err != nil {
		log.Printf("[ERROR] Failed to get user for id %s: %v", userId, err)
		resp.WriteHeader(http.StatusInternalServerError)
		resp.Write([]byte(`{"success": false, "message": "Internal server error"}`))
		return
	}
	if user == nil {
		log.Printf("[ERROR] User with id %s not found", userId)
		resp.WriteHeader(http.StatusNotFound)
		resp.Write([]byte(`{"success": false, "message": "User not found"}`))
		return
	}

	var requestBody struct {
		Otp string `json:"otp"`
	}
	if err := json.NewDecoder(req.Body).Decode(&requestBody); err != nil {
		resp.WriteHeader(http.StatusBadRequest)
		resp.Write([]byte(`{"success": false, "message": "Invalid request body"}`))
		return
	}

	cacheKey := fmt.Sprintf("email_otp_%s", userId)
	cachedItem, exists := repo.GetCache(cacheKey)
	if !exists {
		resp.WriteHeader(http.StatusGone)
		resp.Write([]byte(`{"success": false, "message": "OTP expired"}`))
		return
	}
	cachedOtp := cachedItem.(models.CacheItem).Value.(string)
	if cachedOtp != requestBody.Otp {
		resp.WriteHeader(http.StatusBadRequest)
		resp.Write([]byte(`{"success": false, "message": "Invalid OTP"}`))
		return
	}

	user.EmailVerified = true
	if (user.XVerified || user.LinkedinVerified) && user.HashnodeVerified && user.EmailVerified {
		user.Verified = true
	} else {
		user.Verified = false
	}
	if err := repo.UpdateUser(userId, user); err != nil {
		log.Printf("[ERROR] Failed to update user with id %s: %v", userId, err)
		resp.WriteHeader(http.StatusInternalServerError)
		resp.Write([]byte(`{"success": false, "message": "Internal server error"}`))
		return
	}

	//delete the otp from the cache
	if err := repo.DeleteCache(cacheKey); err != nil {
		log.Printf("[ERROR] Failed to delete OTP from cache for user with id %s: %v", userId, err)
	}

	log.Printf("[INFO] User with ID %s verified email successfully", userId)
	resp.WriteHeader(http.StatusOK)
	resp.Write([]byte(`{"success": true, "message": "Email verified successfully"}`))
}

func ResetEmailOtpHandler(resp http.ResponseWriter, req *http.Request) {
	userId, ok := req.Context().Value(middlewares.UserIDKey).(string)
	if !ok {
		http.Error(resp, "Unauthorized: User ID not found", http.StatusUnauthorized)
		return
	}
	user, err := repo.GetUserById(userId)
	if err != nil {
		log.Printf("[ERROR] Failed to get user for the id: %s and error is %s", userId, err)
		http.Error(resp, "Internal server error", http.StatusInternalServerError)
		return
	}
	if user == nil {
		log.Printf("[ERROR] User with id: %s not found", userId)
		http.Error(resp, "User not found", http.StatusNotFound)
		return
	}
	// delete the old otp
	cacheKey := fmt.Sprintf("email_otp_%s", userId)
	err = repo.DeleteCache(cacheKey)
	if err != nil {
		log.Printf("[ERROR] Failed to delete old OTP for the user id: %s and error is %s", userId, err)
	}
	// generate new otp
	otp := services.GenerateOTP()
	err = repo.SetCache(cacheKey, otp, 30*time.Minute)
	if err != nil {
		log.Printf("[ERROR] Failed to store new OTP for the user id: %s and error is %s", userId, err)
		http.Error(resp, "Internal server error", http.StatusInternalServerError)
		return
	}

	message := fmt.Sprintf("Your OTP is: %s\n OTP will expire in 30 minutes", otp)

	// this is just a temporary work around to use the existing heap based queue system
	// to send emails asynchronously and we need a better way to do this in the future

	emailTask := models.ScheduledBlogData{}
	shareTime := models.ScheduledBlog{}
	shareTime.ScheduledTime = time.Now()
	emailTask.EmailId = user.UserName
	emailTask.Message = message
	emailTask.ScheduledBlog = shareTime
	emailTask.UserID = userId

	err = taskScheduler.AddTask(emailTask)
	if err != nil {
		log.Printf("[ERROR] Failed adding email task to the queue, reason: %s", err)
		resp.WriteHeader(http.StatusInternalServerError)
		resp.Write([]byte(`{"success" : false}`))
	}

	log.Printf("[INFO] New Email verification OTP generated for the user with ID %s", userId)
	resp.WriteHeader(http.StatusOK)
	resp.Write([]byte(`{"success": true}`))
}

func ForgotPasswordHandler(resp http.ResponseWriter, req *http.Request) {
	var requestBody struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(req.Body).Decode(&requestBody); err != nil {
		http.Error(resp, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := repo.GetUserByName(requestBody.Email)
	if err != nil {
		log.Println("[ERROR] Failed to retrieve user:", err)
		http.Error(resp, "Internal server error", http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(resp, "User not found", http.StatusNotFound)
		return
	}
	userId := user.Id.Hex()

	otp := services.GenerateOTP()
	hashedOtp, err := bcrypt.GenerateFromPassword([]byte(otp), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("[ERROR] Failed to hash OTP for user %s: %v", user.UserName, err)
		http.Error(resp, "Internal server error", http.StatusInternalServerError)
		return
	}

	cacheKey := fmt.Sprintf("password_reset_%s", userId)
	if err = repo.DeleteCache(cacheKey); err != nil {
		log.Printf("[WARNING] Failed to delete old OTP for user %s: %v", user.UserName, err)
	}

	if err = repo.SetCache(cacheKey, string(hashedOtp), 10*time.Minute); err != nil {
		log.Printf("[ERROR] Failed to store OTP for user %s: %v", user.UserName, err)
		http.Error(resp, "Internal server error", http.StatusInternalServerError)
		return
	}

	message := fmt.Sprintf("Your OTP for password reset is: %s\n(Expires in 10 minutes)", otp)

	// this is just a temporary work around to use the existing heap based queue system
	// to send emails asynchronously and we need a better way to do this in the future

	emailTask := models.ScheduledBlogData{}
	shareTime := models.ScheduledBlog{}
	shareTime.ScheduledTime = time.Now()
	emailTask.EmailId = user.UserName
	emailTask.Message = message
	emailTask.ScheduledBlog = shareTime
	emailTask.UserID = userId

	err = taskScheduler.AddTask(emailTask)
	if err != nil {
		log.Printf("[ERROR] Failed adding email task to the queue, reason: %s", err)
		resp.WriteHeader(http.StatusInternalServerError)
		resp.Write([]byte(`{"success" : false}`))
	}

	log.Printf("[INFO] Password reset OTP sent successfully for user %s", user.UserName)
	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(http.StatusOK)
	resp.Write([]byte(`{"success": true}`))
}

func ResetPasswordHandler(resp http.ResponseWriter, req *http.Request) {
	var requestBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Otp      string `json:"otp"`
	}
	if err := json.NewDecoder(req.Body).Decode(&requestBody); err != nil {
		http.Error(resp, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(requestBody.Password) < 8 || len(requestBody.Password) > 64 {
		http.Error(resp, "Password must be between 8 and 64 characters", http.StatusBadRequest)
		return
	}

	user, err := repo.GetUserByName(requestBody.Email)
	if err != nil {
		log.Println("[ERROR] Failed to retrieve user:", err)
		http.Error(resp, "Internal server error", http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(resp, "User not found", http.StatusNotFound)
		return
	}
	userId := user.Id.Hex()
	cacheKey := fmt.Sprintf("password_reset_%s", userId)
	cachedData, exists := repo.GetCache(cacheKey)
	if !exists {
		log.Printf("[ERROR] OTP expired or missing for user %s", userId)
		http.Error(resp, "OTP expired", http.StatusGone)
		return
	}

	cachedOtp, ok := cachedData.(models.CacheItem).Value.(string)
	if !ok {
		log.Printf("[ERROR] OTP stored in cache has invalid format for user %s", userId)
		http.Error(resp, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err = bcrypt.CompareHashAndPassword([]byte(cachedOtp), []byte(requestBody.Otp)); err != nil {
		log.Printf("[ERROR] Invalid OTP provided for user %s", userId)
		http.Error(resp, "Invalid OTP", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(requestBody.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("[ERROR] Failed to hash password for user %s: %v", userId, err)
		http.Error(resp, "Internal server error", http.StatusInternalServerError)
		return
	}

	user.PassWord = string(hashedPassword)
	if err = repo.UpdateUser(userId, user); err != nil {
		log.Printf("[ERROR] Failed to update password for user %s: %v", userId, err)
		http.Error(resp, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err = repo.DeleteCache(cacheKey); err != nil {
		log.Printf("[WARNING] Failed to delete OTP for user %s: %v", userId, err)
	}

	log.Printf("[INFO] Password reset successfully for user %s", userId)
	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(http.StatusOK)
	resp.Write([]byte(`{"success": true}`))
}

func DeleteAccountHandler(resp http.ResponseWriter, req *http.Request) {
	userId, ok := req.Context().Value(middlewares.UserIDKey).(string)
	if !ok {
		resp.WriteHeader(http.StatusUnauthorized)
		resp.Write([]byte(`{"success": false, "reason": "Unauthorized: User ID not found"}`))
		return
	}

	type deleteAccountRequest struct {
		Password string `json:"password"`
	}

	var reqBody deleteAccountRequest
	if err := json.NewDecoder(req.Body).Decode(&reqBody); err != nil {
		resp.WriteHeader(http.StatusBadRequest)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	if reqBody.Password == "" {
		resp.WriteHeader(http.StatusBadRequest)
		resp.Write([]byte(`{"success": false, "reason": "Password is required"}`))
		return
	}

	user, err := repo.GetUserById(userId)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			resp.WriteHeader(http.StatusNotFound)
			resp.Write([]byte(`{"success": false, "reason": "User not found"}`))
			return
		}
		log.Printf("[ERROR] Failed to get user for id %s: %s", userId, err)
		resp.WriteHeader(http.StatusInternalServerError)
		resp.Write([]byte(`{"success": false, "reason": "Internal server error"}`))
		return
	}

	// Check if the password matches
	err = bcrypt.CompareHashAndPassword([]byte(user.PassWord), []byte(reqBody.Password))
	if err != nil {
		resp.WriteHeader(http.StatusUnauthorized)
		resp.Write([]byte(`{"success": false, "reason": "Incorrect password"}`))
		return
	}

	// Delete any scheduled tasks before cache cleanup
	for _, blog := range user.ScheduledBlogs {
		if err := taskScheduler.RemoveTask(blog.Id); err != nil {
			log.Printf("[ERROR] Failed to remove scheduled task %s: %s", blog.Id, err)
		}
	}

	cookie, err := req.Cookie("session_token")
	if err == nil {
		if err := repo.DeleteCache(cookie.Value); err != nil {
			log.Printf("[ERROR] Failed to delete session for user %s: %s", userId, err)
		}
	}

	// delete csrf tokens related to the user
	cacheKey := fmt.Sprintf("CSRF_%s", userId)
	err = repo.DeleteCache(cacheKey)
	if err != nil {
		log.Printf("[ERROR] Failed to delete CSRF token for user %s: %s", userId, err)
	}

	// Clear session cookie from client
	http.SetCookie(resp, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	// Delete other user-related cache entries
	cacheKeys := []string{
		fmt.Sprintf("email_otp_%s", userId),
		fmt.Sprintf("password_reset_%s", userId),
	}

	for _, cacheKey := range cacheKeys {
		if err := repo.DeleteCache(cacheKey); err != nil {
			log.Printf("[WARN] Failed to delete cache %s for user %s: %s", cacheKey, userId, err)
		}
	}

	// Delete the user from the database
	err = repo.DeleteUserById(userId)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			resp.WriteHeader(http.StatusNotFound)
			resp.Write([]byte(`{"success": false, "reason": "User not found"}`))
			return
		}
		log.Printf("[ERROR] Failed to delete user %s: %s", userId, err)
		resp.WriteHeader(http.StatusInternalServerError)
		resp.Write([]byte(`{"success": false, "reason": "Internal server error"}`))
		return
	}

	log.Printf("[INFO] User %s deleted successfully", userId)
	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(http.StatusOK)
	resp.Write([]byte(`{"success": true}`))
}
