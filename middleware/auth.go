package middleware

import (
	"context"
	"net"
	"net/http"

	"github.com/MRegterschot/GbxConnector/lib"
)

// Context key for storing the authenticated user
type contextKey string

const UserContextKey = contextKey("user")

// RequireRoles returns a middleware that checks for valid JWT and required roles
func RequireRoles(admin bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !admin {
				// If not admin, just pass through without checking token
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			token := lib.ExtractBearerToken(authHeader)

			// Allow requests from localhost or internal Docker network without token
			remoteIP := r.RemoteAddr
			host, _, err := net.SplitHostPort(remoteIP)
			if err != nil {
				host = remoteIP // fallback if no port
			}

			if host == "127.0.0.1" || host == "::1" || lib.IsDockerInternalIP(host) {
				next.ServeHTTP(w, r)
				return
			}

			// If token not in header, try query param
			if token == "" {
				token = r.URL.Query().Get("token")
			}

			user, err := lib.ValidateAndGetUser(token)
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			if !user.Admin {
				http.Error(w, "Forbidden: insufficient permissions", http.StatusForbidden)
				return
			}

			// Store the user in the context for later use
			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
