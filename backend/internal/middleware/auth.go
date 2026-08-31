package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"attendance-backend/internal/auth"
	"attendance-backend/internal/authz"
	"attendance-backend/internal/httpapi"
)

type ctxKey string

const actorCtxKey ctxKey = "actor"

// RequireAuth validates the bearer access token and loads a fresh Actor
// (role + activity assignments) from the database on every request. The
// actor is never derived from client-supplied data.
func RequireAuth(tm *auth.TokenManager, pool *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				httpapi.Error(w, http.StatusUnauthorized, "missing bearer token")
				return
			}
			token := strings.TrimPrefix(header, "Bearer ")

			claims, err := tm.ParseAccessToken(token)
			if err != nil {
				httpapi.Error(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			actor, err := authz.LoadActor(r.Context(), pool, claims.UserID)
			if err != nil {
				httpapi.Error(w, http.StatusUnauthorized, "account not found")
				return
			}
			if !actor.IsActive {
				httpapi.Error(w, http.StatusForbidden, "account is deactivated")
				return
			}

			ctx := context.WithValue(r.Context(), actorCtxKey, actor)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func ActorFromContext(ctx context.Context) *authz.Actor {
	actor, _ := ctx.Value(actorCtxKey).(*authz.Actor)
	return actor
}

// RequireAdmin gates a route to admin users only. Coach-facing routes
// should NOT use this — they rely on the finer-grained authz guards
// applied inside each service method instead.
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actor := ActorFromContext(r.Context())
		if actor == nil || !actor.IsAdmin() {
			httpapi.Error(w, http.StatusForbidden, "admin access required")
			return
		}
		next.ServeHTTP(w, r)
	})
}
