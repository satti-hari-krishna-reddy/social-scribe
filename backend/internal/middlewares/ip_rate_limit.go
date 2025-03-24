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
			if services.IsIPRateLimited(r, r.URL.Path, limit, duration) {
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"success": false, "reason": "Rate limit exceeded"}`))
				return
			}

			w.Header().Set("Content-Security-Policy", "frame-ancestors 'self'")
			w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=(), usb=()")
			w.Header().Set("Access-Control-Expose-Headers", "X-Csrf-Token") // telling the browser to allow this custom header

			next.ServeHTTP(w, r)
		})
	}
}
