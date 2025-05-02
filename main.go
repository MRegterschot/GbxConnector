package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MRegterschot/GbxConnector/app"
	"github.com/MRegterschot/GbxConnector/config"
	"github.com/MRegterschot/GbxConnector/structs"
	"go.uber.org/zap"
)

func main() {
	srv, err := app.SetupAndRunApp()
	if err != nil {
		zap.L().Fatal("App setup failed", zap.Error(err))
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop

	zap.L().Info("Received shutdown signal, shutting down...")

	ShutdownServers(config.AppEnv.Servers)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		zap.L().Error("HTTP server shutdown failed", zap.Error(err))
	} else {
		zap.L().Info("HTTP server shutdown complete")
	}
}

func ShutdownServers(servers []*structs.Server) {
	for _, server := range servers {
		if server.CancelFunc != nil {
			zap.L().Info("Shutting down server", zap.String("host", server.Host), zap.Int("port", server.XMLRPCPort))
			server.CancelFunc()
		}
	}
}
