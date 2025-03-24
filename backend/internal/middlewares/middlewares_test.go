package middlewares

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"social-scribe/backend/internal/models"
	"social-scribe/backend/internal/repositories"
	"social-scribe/backend/internal/services"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Create a fixed ObjectID for rate-limited users.
var rateLimitedID, _ = primitive.ObjectIDFromHex("000000000000000000000001")

// mockGetCache simulates session lookup.
// It returns a valid session for "valid_token" and "rate_limited_token".
// For "rate_limited_token", we return a CacheItem with a fixed ObjectID.
func mockGetCache(key string) (interface{}, bool) {
	switch key {
	case "valid_token":
		return models.CacheItem{
			Value:     primitive.NewObjectID(), // Random valid user ID
			ExpiresAt: time.Now().Add(10 * time.Minute),
		}, true
	case "rate_limited_token":
		return models.CacheItem{
			Value:     rateLimitedID, // Fixed user ID that triggers rate limiting
			ExpiresAt: time.Now().Add(10 * time.Minute),
		}, true
	case "expired_token":
		return models.CacheItem{
			Value:     primitive.NewObjectID(),
			ExpiresAt: time.Now().Add(-10 * time.Minute),
		}, true
	default:
		return nil, false
	}
}

// mockIsRateLimited simulates user-based rate limiting.
// It returns true if the userID equals the hex representation of rateLimitedID.
func mockIsRateLimited(userID string, path string, limit int, duration time.Duration) bool {
	return userID == rateLimitedID.Hex()
}

// mockIsIPRateLimited simulates IP-based rate limiting.
// It returns true if RemoteAddr equals "192.168.1.100:1234".
func mockIsIPRateLimited(r *http.Request, path string, limit int, duration time.Duration) bool {
	return r.RemoteAddr == "192.168.1.100:1234"
}

func TestAuthMiddleware(t *testing.T) {
	// Override repository functions for testing.
	// (Requires that repositories.GetCache and repositories.IsRateLimited are variables.)
	repositories.GetCache = mockGetCache
	repositories.IsRateLimited = mockIsRateLimited

	tests := []struct {
		name       string
		token      string // session_token cookie value
		wantStatus int
	}{
		{"Valid session", "valid_token", http.StatusOK},
		{"Missing session token", "", http.StatusUnauthorized},
		{"Expired session", "expired_token", http.StatusUnauthorized},
		{"Invalid token", "invalid_token", http.StatusUnauthorized},
		{"Rate limited", "rate_limited_token", http.StatusTooManyRequests},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/protected", nil)
			// Set a dummy URL so r.URL is non-nil.
			req.URL, _ = url.Parse("/protected")
			if tt.token != "" {
				req.AddCookie(&http.Cookie{Name: "session_token", Value: tt.token})
			}

			rec := httptest.NewRecorder()
			// Create AuthMiddleware with limit 5 requests per 10 seconds.
			handler := AuthMiddleware(5, 10*time.Second, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// If AuthMiddleware passes, return 200 OK.
				w.WriteHeader(http.StatusOK)
			}))

			handler.ServeHTTP(rec, req)
			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

func TestIPRateLimitMiddleware(t *testing.T) {
	// Override the IP rate limiting function in services.
	services.IsIPRateLimited = mockIsIPRateLimited

	tests := []struct {
		name       string
		remoteAddr string
		wantStatus int
	}{
		{"Allowed request", "192.168.1.1:5678", http.StatusOK},
		{"Rate limited request", "192.168.1.100:1234", http.StatusTooManyRequests},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api", nil)
			req.RemoteAddr = tt.remoteAddr

			rec := httptest.NewRecorder()
			handler := IPRateLimitMiddleware(5, 10*time.Second)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			handler.ServeHTTP(rec, req)
			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}
