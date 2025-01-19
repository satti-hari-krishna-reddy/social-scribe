package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"social-scribe/backend/api/v1"
	repo "social-scribe/backend/internal/repositories"

	"github.com/rs/cors"
)

func setupCors() *cors.Cors {
	return cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://192.168.29.3:9696", "http://192.168.29.3:5173"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS", "PUT", "DELETE", "PATCH"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "X-Requested-With"},
		AllowCredentials: true,
	})
}

func main() {
	repo.InitMongoDb()
	router := v1.RegisterRoutes()
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "MISSING"
	}

	corsHandler := setupCors()

	port := os.Getenv("BACKEND_PORT")
	if port == "" {
		log.Printf("[DEBUG] Running on %s:9696", hostname)
		log.Fatal(http.ListenAndServe(":9696", corsHandler.Handler(router)))
	} else {
		log.Printf("[DEBUG] Running on %s:%s", hostname, port)
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), corsHandler.Handler(router)))
	}
}

// package main

// import (
// 	"bytes"
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"io/ioutil"
// 	"log"
// 	"net/http"
// 	"strings"

// 	// "os"

// 	"github.com/dghubble/oauth1"
// 	"github.com/dghubble/oauth1/twitter"
// 	"github.com/gorilla/mux"

// 	// "github.com/joho/godotenv"
// 	"github.com/google/uuid"
// 	"github.com/rs/cors"
// 	"golang.org/x/oauth2"
// 	"golang.org/x/oauth2/linkedin"
// )

// type TweetRequest struct {
// 	Tweet string `json:"tweet"`
// }

// type GraphQLQuery struct {
// 	Query string `json:"query"`
// }

// type CoverImage struct {
// 	URL string `json:"url"`
// }

// type Author struct {
// 	Name string `json:"name"`
// }

// type PostNode struct {
// 	Title            string     `json:"title"`
// 	URL              string     `json:"url"`
// 	ID               string     `json:"id"`
// 	CoverImage       CoverImage `json:"coverImage"`
// 	Author           Author     `json:"author"`
// 	ReadTimeInMinutes int        `json:"readTimeInMinutes"`
// }

// type Edge struct {
// 	Node PostNode `json:"node"`
// }

// type Posts struct {
// 	Edges []Edge `json:"edges"`
// }

// type Publication struct {
// 	Posts Posts `json:"posts"`
// }

// type Data struct {
// 	Publication Publication `json:"publication"`
// }

// type GraphQLResponse struct {
// 	Data Data `json:"data"`
// }

// var (
// 	config     *oauth1.Config
// 	linkedinConfig  *oauth2.Config
// 	tokenCache map[string]*oauth1.Token // Temporary in-memory cache; replace with DB in production
// )
// // var linkedinConfig = &oauth2.Config{
// // 	ClientID:     string
// // 	ClientSecret: string
// // 	RedirectURL:  string
// // 	Scopes:       []string
// // }
// func init() {
// 	// // Load environment variables
// 	// err := godotenv.Load()
// 	// if err != nil {
// 	// 	log.Fatal("Error loading .env file")
// 	// }

// 	config = &oauth1.Config{
// 		ConsumerKey:    "QcrVJFLQkCRptSoibPjPctQjB",
// 		ConsumerSecret: "JSdAV5PH7CYf9qHxGmkhdUqvJ4PrqtvhXry6ztmqbWRAQJkGfd",
// 		CallbackURL:    "http://localhost:8080/callback",
// 		Endpoint:       twitter.AuthorizeEndpoint,
// 	}

// 	linkedinConfig = &oauth2.Config{
// 		ClientID:     "86vtgt68dgno8k",
// 		ClientSecret:  "WPL_AP1.bFL6CSp3rVrDMRrB.Y8NYfw==",
// 		RedirectURL:  "http://localhost:8080/linkedin-callback",
// 		Scopes:       []string{"w_member_social"},
// 		Endpoint:     linkedin.Endpoint,
// 	}

// 	tokenCache = make(map[string]*oauth1.Token)
// }

// func main() {
// 	r := mux.NewRouter()
// 	r.HandleFunc("/connect-twitter", connectTwitterHandler).Methods(http.MethodGet, http.MethodPost, http.MethodOptions, http.MethodHead)
// 	r.HandleFunc("/callback", XcallbackHandler).Methods(http.MethodGet, http.MethodPost, http.MethodOptions)
// 	r.HandleFunc("/post-tweet", postTweetHandler).Methods(http.MethodPost, http.MethodOptions)
// 	r.HandleFunc("/connect-linkedin", connectLinkedInHandler).Methods(http.MethodGet, http.MethodPost, http.MethodOptions, http.MethodHead)
// 	r.HandleFunc("/linkedin-callback", LinkedCallbackHandler).Methods(http.MethodGet, http.MethodPost, http.MethodOptions)
// 	r.HandleFunc("/post-linkedin", LinkedPostHandler).Methods(http.MethodPost, http.MethodOptions)

// 	corsHandler := setupCors()

// 	// Start the server with CORS middleware applied
// 	log.Printf("Starting server on port 8080...")
// 	log.Fatal(http.ListenAndServe(":8080", corsHandler.Handler(r)))
// }

// func fetchPublicationPosts(host string) ([]Edge, error) {
// 	// Define the GraphQL endpoint
// 	endpoint := "https://gql.hashnode.com"

// 	// Define the GraphQL query
// 	query := GraphQLQuery{
// 		Query: fmt.Sprintf(`
// 			query Publication {
// 			  publication(host: \"%s\") {
// 			    posts(first: 0) {
// 			      edges {
// 			        node {
// 			          title
// 			          url
// 			          id
// 			          coverImage {
// 			            url
// 			          }
// 			          author {
// 			            name
// 			          }
// 			          readTimeInMinutes
// 			        }
// 			      }
// 			    }
// 			  }
// 			}
// 		`, host),
// 	}

// 	// Convert the query to JSON
// 	queryBytes, err := json.Marshal(query)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to marshal query: %v", err)
// 	}

// 	// Make the HTTP POST request to the GraphQL endpoint
// 	headers := map[string]string{
// 		"Content-Type": "application/json",
// 		// "Authorization": "Bearer YOUR_ACCESS_TOKEN",
// 	}

// 	response, err := makePostRequest(endpoint, queryBytes, headers)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to make GraphQL request: %v", err)
// 	}

// 	// Parse the JSON response
// 	var gqlResponse GraphQLResponse
// 	err = json.Unmarshal(response, &gqlResponse)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to parse response: %v", err)
// 	}

// 	// Extract the edges from the response
// 	edges := gqlResponse.Data.Publication.Posts.Edges
// 	return edges, nil
// }

// func makePostRequest(url string, body []byte, headers map[string]string) ([]byte, error) {
// 	request, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create HTTP request: %v", err)
// 	}

// 	for key, value := range headers {
// 		request.Header.Set(key, value)
// 	}

// 	client := &http.Client{}
// 	response, err := client.Do(request)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to execute HTTP request: %v", err)
// 	}
// 	defer response.Body.Close()

// 	if response.StatusCode != http.StatusOK {
// 		body, _ := ioutil.ReadAll(response.Body)
// 		return nil, fmt.Errorf("GraphQL query failed with status code %d: %s", response.StatusCode, string(body))
// 	}

// 	return ioutil.ReadAll(response.Body)
// }
