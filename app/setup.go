package app

import (
	"context"
	"net/http"
	"strconv"

	"github.com/MRegterschot/GbxConnector/config"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func SetupAndRunApp() error {
	// load env
	err := config.LoadEnv()
	if err != nil {
		return err
	}

	config.SetupLogger()

	for _, server := range config.AppEnv.Servers {
		GetClient(server)

		// Call cancel() to stop the reconnect loop when the context is done
		ctx, cancel := context.WithCancel(context.Background())
		StartReconnectLoop(ctx, server)
		defer cancel()
	}

	if len(config.AppEnv.CorsOrigins) == 0 {
		zap.L().Warn("No CORS origins configured, allowing all origins")
	}

	// Create a new Gorilla Mux router
	router := mux.NewRouter()

	// Set up routes
	SetupRoutes(router)

	// Attach middleware
	handler := loggingMiddleware(recoveryMiddleware(router))

	port := strconv.Itoa(config.AppEnv.Port)
	zap.L().Info("Starting server", zap.String("port", port))
	return http.ListenAndServe(":"+port, handler)
}
