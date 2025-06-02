package app

import (
	"net/http"
	"time"

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
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Preflight request
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
