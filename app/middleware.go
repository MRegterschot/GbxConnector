package app

import (
	"net/http"
	"time"

	"slices"

	"github.com/MRegterschot/GbxConnector/config"
	"go.uber.org/zap"
)

// Logging middleware to log requests using Zap
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Log the incoming request using Zap
		zap.L().Debug("Received request",
			zap.String("remote_addr", r.RemoteAddr),
			zap.String("method", r.Method),
			zap.String("url", r.URL.String()),
			zap.String("user_agent", r.UserAgent()),
		)

		// Call the next handler
		next.ServeHTTP(w, r)

		// Log the response time after processing
		zap.L().Debug("Request processed",
			zap.String("remote_addr", r.RemoteAddr),
			zap.String("method", r.Method),
			zap.String("url", r.URL.String()),
			zap.Duration("latency", time.Since(start)),
		)
	})
}

// Recovery middleware to handle panic and prevent server crash
func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				// Log the panic error
				zap.L().Error("Recovered from panic", zap.Any("error", r))

				// Respond with a generic internal server error message
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowedOrigins := config.AppEnv.CorsOrigins

		allowed := slices.Contains(allowedOrigins, origin)

		zap.L().Debug("CORS request",
			zap.String("origin", origin),
			zap.String("method", r.Method),
			zap.String("url", r.URL.String()),
			zap.Bool("allowed", allowed),
		)

		// If the Origin header is present and not allowed, deny access
		if origin != "" && !allowed {
			http.Error(w, "Forbidden - CORS origin denied", http.StatusForbidden)
			return
		}

		// Set CORS headers only if allowed
		if allowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		}

		// Preflight request
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
