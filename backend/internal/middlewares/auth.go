package middlewares

import (
	"context"
	"net/http"
	"time"

	repo "social-scribe/backend/internal/repositories"
	"social-scribe/backend/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type contextKey string

const UserIDKey contextKey = "userID"

// AuthMiddleware handles authentication and rate limiting
func AuthMiddleware(limit int, duration time.Duration, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Validate session
		cookie, err := r.Cookie("session_token")
		if err != nil {
			http.Error(w, "Unauthorized: Missing session token", http.StatusUnauthorized)
			return
		}

		sessionData, exists := repo.GetCache(cookie.Value)
		if !exists {
			http.Error(w, "Unauthorized: Invalid or expired session", http.StatusUnauthorized)
			return
		}

		session, ok := sessionData.(models.CacheItem)
		if !ok {
			http.Error(w, "Unauthorized: Invalid session format", http.StatusUnauthorized)
			return
		}

		if session.ExpiresAt.Before(time.Now()) {
			http.Error(w, "Unauthorized: Session expired", http.StatusUnauthorized)
			return
		}

		oid, ok := session.Value.(primitive.ObjectID)
		if !ok {
			http.Error(w, "Unauthorized: Invalid user ID format", http.StatusUnauthorized)
			return
		}

		userID := oid.Hex()

		if repo.IsRateLimited(userID, limit, duration) {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
