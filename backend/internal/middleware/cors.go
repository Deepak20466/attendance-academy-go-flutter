package middleware

import (
	"net/http"

	"github.com/go-chi/cors"
)

// CORS is permissive on origin because the API authenticates with a Bearer
// token in the Authorization header, never cookies — there is no
// credentialed cross-site request to protect against, so a wildcard origin
// is safe here (browsers refuse to combine AllowCredentials with "*", and
// we don't need credentials mode at all).
func CORS() func(http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	})
}
