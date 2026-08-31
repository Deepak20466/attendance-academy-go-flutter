package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/httprate"
)

// General API rate limit: generous, keyed by client IP.
func RateLimit(next http.Handler) http.Handler {
	return httprate.LimitByIP(300, time.Minute)(next)
}

// AuthRateLimit is stricter and applied only to login/refresh endpoints to
// slow down credential-stuffing / brute-force attempts.
func AuthRateLimit(next http.Handler) http.Handler {
	return httprate.LimitByIP(10, time.Minute)(next)
}
