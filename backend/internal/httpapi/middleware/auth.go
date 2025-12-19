package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"resume-tailor/internal/auth"
	"resume-tailor/internal/httpapi/cookies"
)

// ctxKey is a private type for context keys to avoid collisions.
type ctxKey string

const userIDKey ctxKey = "userID"

// WithUserID injects a user ID into the request context.
func WithUserID(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey, id)
}

// UserIDFromContext extracts the user ID from the request context.
// Returns the user ID and true if found, or uuid.Nil and false otherwise.
func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(userIDKey).(uuid.UUID)
	if !ok {
		return uuid.Nil, false
	}
	return id, true
}

// AuthRequired is a middleware that validates the session cookie and injects the user ID into the context.
// If no valid session cookie is found or authentication fails, it returns a 401 Unauthorized JSON response.
func AuthRequired(authSvc *auth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Read session cookie
			token, ok := cookies.ReadSessionCookie(r)
			if !ok {
				writeUnauthorized(w)
				return
			}

			// Trim whitespace from token
			token = strings.TrimSpace(token)
			if token == "" {
				writeUnauthorized(w)
				return
			}

			// Authenticate the token
			userID, err := authSvc.Authenticate(r.Context(), token)
			if err != nil {
				writeUnauthorized(w)
				return
			}

			// Inject user ID into context and continue
			ctx := WithUserID(r.Context(), userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// writeUnauthorized writes a 401 Unauthorized JSON response.
func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]string{
		"error": "unauthorized",
	})
}

