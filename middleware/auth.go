package middleware

import (
	"context"
	"net/http"

	"github.com/MRegterschot/GbxConnector/config"
	"github.com/MRegterschot/GbxConnector/lib"
)

// Context key for storing the authenticated user
type contextKey string

const UserContextKey = contextKey("user")

// RequireRoles returns a middleware that checks for valid JWT and required roles
func RequireRoles(allowedRoles ...string) func(http.Handler) http.Handler {
	roleSet := make(map[string]bool)
	for _, role := range allowedRoles {
		roleSet[role] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			token := lib.ExtractBearerToken(authHeader)

			
			// If token not in header, try query param
			if token == "" {
				token = r.URL.Query().Get("token")
			}

			if token == "" {
				http.Error(w, "Missing or invalid Authorization token", http.StatusUnauthorized)
				return
			}

			// Bypass for internal API key
			if token == config.AppEnv.InternalApiKey {
				next.ServeHTTP(w, r) // trusted
				return
			}

			user, err := lib.ValidateAndGetUser(token)
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Check if user has at least one of the allowed roles
			hasRole := false
			for _, role := range user.Roles {
				if roleSet[role] {
					hasRole = true
					break
				}
			}

			if !hasRole {
				http.Error(w, "Forbidden: insufficient permissions", http.StatusForbidden)
				return
			}

			// Store the user in the context for later use
			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
