package app

import (
	"context"
	"net/http"
	"strconv"

	"github.com/MRegterschot/GbxConnector/config"
	"github.com/MRegterschot/GbxConnector/handlers"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func SetupAndRunApp() (*http.Server, error) {
	// load env
	err := config.LoadEnv()
	if err != nil {
		return nil, err
	}

	config.SetupLogger()

	// Register handlers
	handlers.SetAddServerFunc(AddServer)
	handlers.SetRemoveServerFunc(DeleteServer)
	handlers.SetUpdateServerFunc(UpdateServer)
	handlers.SetOrderServersFunc(OrderServers)

	zap.L().Info("Found servers", zap.Int("count", len(config.AppEnv.Servers)))
	for _, server := range config.AppEnv.Servers {
		GetClient(server)
		handlers.GetListenerSocket(server.Id)

		ctx, cancel := context.WithCancel(context.Background())
		server.Ctx = ctx
		server.CancelFunc = cancel

		go StartReconnectLoop(ctx, server)
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

	srv := &http.Server{
		Addr:    ":" + strconv.Itoa(config.AppEnv.Port),
		Handler: handler,
	}

	// Run HTTP server in a goroutine
	go func() {
		zap.L().Info("Starting server", zap.String("port", strconv.Itoa(config.AppEnv.Port)))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.L().Fatal("HTTP server failed", zap.Error(err))
		}
	}()

	return srv, nil
}
