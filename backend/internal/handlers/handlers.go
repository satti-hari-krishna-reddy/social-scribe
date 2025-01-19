package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	// "os/user"
	"social-scribe/backend/internal/models"
	repo "social-scribe/backend/internal/repositories"
	"strings"
	"time"

	"github.com/dghubble/oauth1"
	"github.com/dghubble/oauth1/twitter"
	"github.com/golang-jwt/jwt/v5"

	// "github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/linkedin"
)

var twitterConfig = &oauth1.Config{}
var linkedinConfig = &oauth2.Config{}

func init() {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
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
		Scopes:       []string{"w_member_social"},
		Endpoint:     linkedin.Endpoint,
	}

}

var tokenCache = make(map[string]*oauth1.Token)

func SignupUserHandler(resp http.ResponseWriter, req *http.Request) {
	if req.Body == nil {
		http.Error(resp, `{"error": "Failed to parse credintials: body is empty"}`, http.StatusBadRequest)
		return
	}
	user := models.User{}

	err := json.NewDecoder(req.Body).Decode(&user)
	if err != nil {
		http.Error(resp, `{"error": "Bad request: unable to decode JSON"}`, http.StatusBadRequest)
		return
	}

	user.UserName = strings.TrimSpace(user.UserName)
	user.UserName = strings.Join(strings.Fields(strings.ToLower(user.UserName)), "")
	user.PassWord = strings.TrimSpace(user.PassWord)

	if len(user.UserName) < 4 || len(user.UserName) > 64 {
		http.Error(resp, `{"error": "The username should contain a minimum of 4 and maximum of 64 characters"}`, http.StatusBadRequest)
		return
	}
	if len(user.PassWord) < 8 || len(user.PassWord) > 128 {
		http.Error(resp, `{"error": "The password should contain a minimum of 8 and maximun of 128 characters"}`, http.StatusBadRequest)
		return
	}

	existingUser, err := repo.GetUserByName(user.UserName)
	if err != nil {
		http.Error(resp, `{"error": "Internal server error"}`, http.StatusInternalServerError)
		log.Printf("[ERROR] Error checking existing user: %v", err)
		return
	}
	if existingUser != nil {
		http.Error(resp, `{"message" : "Username already taken"}`, http.StatusConflict)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.PassWord), bcrypt.DefaultCost)
	if err != nil {
		http.Error(resp, `{"error": "Internal server error"}`, http.StatusInternalServerError)
		log.Printf("[ERROR] Error hashing password for user '%s': %v", user.UserName, err)
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
		log.Printf("[ERROR] Unable to create user %v: %v", user.UserName, err)
		http.Error(resp, `{"error": "Failed to create user"}`, http.StatusInternalServerError)
		return
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userId,
		"email":   user.UserName,
		"exp":     time.Now().Add(24 * time.Hour).Unix(), // 1-day expiration
	})
	tokenString, err := token.SignedString([]byte("your-secret-key"))
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		resp.Write([]byte(`{"success": false, "reason": "Failed to generate token"}`))
		return
	}
	http.SetCookie(resp, &http.Cookie{
		Name:     "auth_token",
		Value:    tokenString,
		HttpOnly: true,
		Path:     "/",
		Secure:   false,
	})

	user.PassWord = ""
	responseJson, err := json.Marshal(user)
	if err != nil {
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Failed unpacking user"}`))
		return
	}

	log.Printf("[INFO] User '%s' successfully registered with ID: %s", user.UserName, userId)

	resp.WriteHeader(http.StatusCreated)
	resp.Header().Set("Content-Type", "application/json")
	resp.Write([]byte(responseJson))
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

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.Id,
		"email":   user.UserName,
		"exp":     time.Now().Add(24 * time.Hour).Unix(), // 1-day expiration
	})
	tokenString, err := token.SignedString([]byte("your-secret-key"))
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		resp.Write([]byte(`{"success": false, "reason": "Failed to generate token"}`))
		return
	}
	http.SetCookie(resp, &http.Cookie{
		Name:     "auth_token",
		Value:    tokenString,
		HttpOnly: true,
		Path:     "/",
		Secure:   false,
	})

	user.PassWord = ""
	responseJson, err := json.Marshal(user)
	if err != nil {
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Failed unpacking user"}`))
		return
	}

	resp.WriteHeader(http.StatusAccepted)
	resp.Header().Set("Content-Type", "application/json")
	resp.Write([]byte(responseJson))
}

func GetUserInfoHandler(resp http.ResponseWriter, req *http.Request) {
	cookie, err := req.Cookie("auth_token")
	if err != nil {
		http.Error(resp, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}
	token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte("your-secret-key"), nil
	})

	if err != nil {
		http.Error(resp, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		http.Error(resp, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}
	userId, ok := claims["user_id"].(string)
	if !ok {
		http.Error(resp, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
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
	// user.HashnodePAT = ""
	// user.LinkedInOauthKey = ""
	// user.XoauthKey = ""
	responseJson, err := json.Marshal(user)
	if err != nil {
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Failed unpacking"}`))
		return
	}

	resp.WriteHeader(http.StatusOK)
	resp.Write(responseJson)
}

func GetUserNotificationsHandler(resp http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	userId := vars["id"]
	if len(userId) == 0 {
		http.Error(resp, `{"error": "cant able parse id field, reason is missing id field in the request"}`, http.StatusBadRequest)
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

func GetUserSharedBlogsHandler(resp http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	userId := vars["id"]
	if len(userId) == 0 {
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success" : false, "reason" : "user id not found in the request}`))
		return
	}
	user, err := repo.GetUserById(userId)
	if err != nil {
		log.Printf("[ERROR] Failed to find user for the id: %s and error is %s", userId, err)
		resp.WriteHeader(500)
		resp.Write([]byte(`{"success" : false}`))
		return
	}
	if user == nil {
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success" : false, "reason" : "user id is invalid"}`))
		return
	}
	response := map[string]interface{}{
		"shared_blogs": user.SharedBlogs,
	}
	responseJson, err := json.Marshal(response)
	if err != nil {
		resp.WriteHeader(401)
		resp.Write([]byte(`{"sucess" : false, "reason" : "Failed unpacking}`))
		return
	}
	resp.WriteHeader(200)
	resp.Write(responseJson)

}

func GetUserScheduledBlogsHandler(resp http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	userId := vars["id"]
	if len(userId) == 0 {
		resp.WriteHeader(401)
		resp.Write([]byte(`"success" : false, "reason" : "user id not provided`))
		return

	}
	user, err := repo.GetUserById(userId)
	if user == nil {
		resp.WriteHeader(401)
		resp.Write([]byte(`"success" : false, "reason" : "user id is invalid`))
		return
	}
	if err != nil {
		log.Printf("[ERROR] Failed to find user for the id: %s and error is %s", userId, err)
		resp.WriteHeader(500)
		resp.Write([]byte(`{"success" : "false"}`))
		return
	}
	response := map[string]interface{}{
		"scheduled_blogs": user.ScheduledBlogs,
	}
	responseJson, err := json.Marshal(response)
	if err != nil {
		resp.WriteHeader(500)
		resp.Write([]byte(`{"success" : false}`))
		return
	}
	resp.WriteHeader(200)
	resp.Write(responseJson)
}

func ClearUserNotificationsHandler(resp http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	userId := vars["id"]
	if len(userId) == 0 {
		resp.WriteHeader(401)
		resp.Write([]byte(`"success" : false, "reason" : "missing user id in the request"`))
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

func ScheduleUserBlogHandler(resp http.ResponseWriter, req *http.Request) {
	var blogData models.ScheduledBlogData
	decoder := json.NewDecoder(req.Body)
	defer req.Body.Close()

	if err := decoder.Decode(&blogData); err != nil {
		http.Error(resp, "Bad request, failed to parse JSON", http.StatusBadRequest)
		return
	}

	if len(blogData.UserID) == 0 {
		resp.WriteHeader(http.StatusBadRequest)
		resp.Write([]byte(`{"success" : false, "reason" : "no user id found"}`))
		return
	}

	user, err := repo.GetUserById(blogData.UserID)
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		resp.Write([]byte(`{"success" : false}`))
		return
	}

	if user == nil {
		resp.WriteHeader(http.StatusBadRequest)
		resp.Write([]byte(`{"success" : false, "reason" : "user id is not valid"}`))
		return
	}

	if len(blogData.ScheduledBlog.ScheduledTime) == 0 {
		resp.WriteHeader(http.StatusBadRequest)
		resp.Write([]byte(`{"success" : false, "reason" : "scheduled time is missing"}`))
		return
	}

	_, err = time.Parse(time.RFC3339, blogData.ScheduledBlog.ScheduledTime)
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
		resp.Write([]byte(`{"success" : false, "reason" : "invalid scheduled time format, must be RFC3339"}`))
		return
	}

	if err := blogData.ScheduledBlog.Validate(); err != nil {
		http.Error(resp, err.Error(), http.StatusBadRequest)
		return
	}

	durableFunctionURL := "https://<your-function-app>.azurewebsites.net/api/orchestrator"
	reqBody, _ := json.Marshal(blogData)

	durableResp, err := http.Post(durableFunctionURL, "application/json", bytes.NewBuffer(reqBody))
	if err != nil || durableResp.StatusCode != http.StatusOK {
		log.Printf("[DEBUG] Failed to create durable function, reason: %s", err)
		resp.WriteHeader(http.StatusInternalServerError)
		resp.Write([]byte(`{"success": false, "reason": "failed to create a cloud function"}`))
		return
	}

	var instanceID string
	if err := json.NewDecoder(durableResp.Body).Decode(&instanceID); err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	resp.WriteHeader(http.StatusOK)
	resp.Write([]byte("Blog scheduled validated"))
}

func GetUserBlogsHandler(host string) ([]models.Edge, error) {
	// Define the GraphQL endpoint
	endpoint := "https://gql.hashnode.com"

	// Define the GraphQL query
	query := models.GraphQLQuery{
		Query: fmt.Sprintf(`
			query Publication {
			  publication(host: \"%s\") {
			    posts(first: 0) {
			      edges {
			        node {
			          title
			          url
			          id
			          coverImage {
			            url
			          }
			          author {
			            name
			          }
			          readTimeInMinutes
			        }
			      }
			    }
			  }
			}
		`, host),
	}

	// Convert the query to JSON
	queryBytes, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %v", err)
	}

	// Make the HTTP POST request to the GraphQL endpoint
	headers := map[string]string{
		"Content-Type": "application/json",
		// "Authorization": "Bearer YOUR_ACCESS_TOKEN",
	}

	response, err := makePostRequest(endpoint, queryBytes, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to make GraphQL request: %v", err)
	}

	// Parse the JSON response
	var gqlResponse models.GraphQLResponse
	err = json.Unmarshal(response, &gqlResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	// Extract the edges from the response
	edges := gqlResponse.Data.Publication.Posts.Edges
	return edges, nil
}

func makePostRequest(url string, body []byte, headers map[string]string) ([]byte, error) {
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %v", err)
	}

	for key, value := range headers {
		request.Header.Set(key, value)
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(response.Body)
		return nil, fmt.Errorf("GraphQL query failed with status code %d: %s", response.StatusCode, string(body))
	}

	return ioutil.ReadAll(response.Body)
}

func ConnectXhandler(w http.ResponseWriter, r *http.Request) {
	userId, err := ValidateLogin(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	//get the user from the database
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
		http.Error(w, "Failed to get request token", http.StatusInternalServerError)
		return
	}

	user.XoauthKey = oauth1.Token{Token: requestToken, TokenSecret: requestSecret}
	err = repo.UpdateUser(userId, user)
	if err != nil {
		log.Printf("[ERROR] Failed to update user with id: %s and error is %s", userId, err)
		return
	}

	authorizationURL, err := twitterConfig.AuthorizationURL(requestToken)
	if err != nil {
		http.Error(w, "Failed to get authorization URL", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, authorizationURL.String(), http.StatusFound)
}

func XcallbackHandler(w http.ResponseWriter, r *http.Request) {

	userID, err := ValidateLogin(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	user, err := repo.GetUserById(userID)
	if err != nil {
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		log.Printf("[ERROR] Failed to get user for the id: %s and error is %s", userID, err)

		return
	}
	if user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		log.Printf("[ERROR] User with id: %s not found", userID)
		return
	}

	requestTokenData := &user.XoauthKey
	verifier := r.URL.Query().Get("oauth_verifier")
	if verifier == "" {
		http.Error(w, "Missing OAuth verifier", http.StatusBadRequest)
		return
	}
	accessToken, accessSecret, err := twitterConfig.AccessToken(requestTokenData.Token, requestTokenData.TokenSecret, verifier)
	if err != nil {
		http.Error(w, "Failed to get access token", http.StatusInternalServerError)
		return
	}

	user.XoauthKey = oauth1.Token{Token: accessToken, TokenSecret: accessSecret}
	user.XVerified = true
	err = repo.UpdateUser(userID, user)
	if err != nil {
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		log.Printf("[ERROR] Failed to update user with id: %s and error is %s", userID, err)
		return
	}

	fmt.Println("Connected to Twitter!")
	http.Redirect(w, r, "http://localhost:5173?x=connected", http.StatusSeeOther)
}

// Handler to post a tweet
func PostTweetHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := ValidateLogin(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	user, err := repo.GetUserById(userID)
	if err != nil {
		log.Printf("[ERROR] Failed to get user for the id: %s and error is %s", userID, err)
		return
	}
	if user == nil {
		log.Printf("[ERROR] User with id: %s not found", userID)
		return
	}
	if !user.XVerified {
		http.Error(w, "Twitter account not connected", http.StatusForbidden)
		return
	}

	userToken := &user.XoauthKey

	var tweetReq models.TweetRequest
	err = json.NewDecoder(r.Body).Decode(&tweetReq)
	if err != nil || tweetReq.Tweet == "" {
		http.Error(w, "Invalid tweet content", http.StatusBadRequest)
		return
	}

	// Create an OAuth1 HTTP client
	client := twitterConfig.Client(oauth1.NoContext, userToken)

	// Make the Twitter API request to post a tweet
	tweetURL := "https://api.twitter.com/1.1/statuses/update.json"
	resp, err := client.PostForm(tweetURL, map[string][]string{"status": {tweetReq.Tweet}})
	if err != nil {
		http.Error(w, "Failed to post tweet: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		http.Error(w, fmt.Sprintf("Twitter API error: %s", resp.Status), http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, "Tweet posted successfully!")
}

func ConnectLinkedInHandler(w http.ResponseWriter, r *http.Request) {
	userId, err := ValidateLogin(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
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
	state := userId
	user.LinkedInOauthKey = state
	err = repo.UpdateUser(userId, user)
	if err != nil {
		log.Printf("[ERROR] Failed to update user with id: %s and error is %s", userId, err)
		return
	}
	// sending userid as state can be a security risk, this needed to be fixed in the future : )
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		HttpOnly: true,
		Path:    "/",
		Secure:   false,
	})

	authURL := linkedinConfig.AuthCodeURL(state)
	http.Redirect(w, r, authURL, http.StatusFound)
}

func LinkedCallbackHandler(w http.ResponseWriter, r *http.Request) {
	queryState := r.URL.Query().Get("state")
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || stateCookie.Value != queryState {
		http.Error(w, "Invalid state parameter", http.StatusForbidden)
		return
	}
	userId := stateCookie.Value
	user, err := repo.GetUserById(userId)
	if err != nil {
		log.Printf("[ERROR] Failed to get user for the id: %s and error is %s", userId, err)
		return
	}
	if user == nil {
		log.Printf("[ERROR] User with id: %s not found", userId)
		return
	}
	if user.LinkedInOauthKey != queryState {
		http.Error(w, "Invalid state parameter", http.StatusForbidden)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	token, err := linkedinConfig.Exchange(ctx, code)
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}
	user.LinkedInOauthKey = token.AccessToken
	user.LinkedinVerified = true
	err = repo.UpdateUser(userId, user)
	if err != nil {
		log.Printf("[ERROR] Failed to update user with id: %s and error is %s", userId, err)
		return
	}
	fmt.Println("Connected to LinkedIn!")

	// Redirect the user back to the frontend
	http.Redirect(w, r, "http://localhost:5173?linkedin=connected", http.StatusSeeOther)
}

func LinkedPostHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := ValidateLogin(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	user, err := repo.GetUserById(userID)
	if err != nil {
		log.Printf("[ERROR] Failed to get user for the id: %s and error is %s", userID, err)
		return
	}
	if user == nil {
		log.Printf("[ERROR] User with id: %s not found", userID)
		return
	}
	if !user.LinkedinVerified {
		http.Error(w, "LinkedIn account not connected", http.StatusForbidden)
		return
	}
	accessToken := user.LinkedInOauthKey
	message := "Hello, LinkedIn!"
	postContent := map[string]interface{}{
		"author":         "urn:li:person:~", // This refers to the authenticated user
		"lifecycleState": "PUBLISHED",
		"specificContent": map[string]interface{}{
			"com.linkedin.ugc.ShareContent": map[string]interface{}{
				"shareCommentary": map[string]interface{}{
					"text": message,
				},
				"shareMediaCategory": "NONE",
			},
		},
		"visibility": map[string]interface{}{
			"com.linkedin.ugc.MemberNetworkVisibility": "PUBLIC",
		},
	}

	client := &http.Client{}
	reqBody, _ := json.Marshal(postContent)
	req, err := http.NewRequest("POST", "https://api.linkedin.com/v2/ugcPosts", strings.NewReader(string(reqBody)))
	if err != nil {
		http.Error(w, "Failed to create post request: "+err.Error(), http.StatusInternalServerError)
		return
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to post content: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		fmt.Fprintln(w, "Post created successfully!")
	} else {
		http.Error(w, "Failed to create post", resp.StatusCode)
	}
}

func ValidateLogin(req *http.Request) (string, error) {
	cookie, err := req.Cookie("auth_token")
	if err != nil {
		return "", err
	}
	token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte("your-secret-key"), nil
	})

	if err != nil {
		return "", err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", fmt.Errorf("invalid token")
	}
	userId, ok := claims["user_id"].(string)
	if !ok {
		return "", fmt.Errorf("invalid user id")
	}
	return userId, nil
}
