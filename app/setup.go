package app

import (
	"context"
	"net/http"
	"strconv"

	"github.com/MRegterschot/GbxConnector/config"
	"github.com/MRegterschot/GbxConnector/handlers"
	"github.com/MRegterschot/GbxConnector/lib"
	"github.com/google/uuid"
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

	go func() {
		zap.L().Info("Found servers", zap.Int("count", len(config.AppEnv.Servers)))
		for _, server := range config.AppEnv.Servers {
			if server.Uuid == "" {
				server.Uuid = uuid.NewString()
			}

			GetClient(server)
			handlers.GetMapSocket(server.Uuid)
			handlers.GetPlayersSocket(server.Uuid)
			handlers.GetLiveSocket(server.Uuid)

			ctx, cancel := context.WithCancel(context.Background())
			server.Ctx = ctx
			server.CancelFunc = cancel

			go StartReconnectLoop(ctx, server)
		}

		// Save servers to ensure UUIDs are set
		if err := lib.WriteFile("./servers.json", &config.AppEnv.Servers); err != nil {
			zap.L().Error("Failed to write servers.json", zap.Error(err))
		}
	}()

	// Create a new Gorilla Mux router
	router := mux.NewRouter()

	// Set up routes
	SetupRoutes(router)

	// Attach middleware
	handler := loggingMiddleware(recoveryMiddleware(corsMiddleware(router)))

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
