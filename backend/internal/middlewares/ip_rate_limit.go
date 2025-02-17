package middlewares

import (
	"net/http"
	"social-scribe/backend/internal/services"
	"time"
)

// IPRateLimitMiddleware applies rate limiting per IP for public routes
func IPRateLimitMiddleware(limit int, duration time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if services.IsIPRateLimited(r, limit, duration) {
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"success": false, "reason": "Rate limit exceeded"}`))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
