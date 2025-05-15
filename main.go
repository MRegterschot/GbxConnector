package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MRegterschot/GbxConnector/app"
	"github.com/MRegterschot/GbxConnector/config"
	"github.com/MRegterschot/GbxConnector/lib"
	"go.uber.org/zap"
)

func main() {
	srv, err := app.SetupAndRunApp()
	if err != nil {
		zap.L().Fatal("App setup failed", zap.Error(err))
	}

	go lib.MemoryChecker(5 * time.Minute)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop

	zap.L().Info("Received shutdown signal, shutting down...")

	app.ShutdownServers(config.AppEnv.Servers)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		zap.L().Error("HTTP server shutdown failed", zap.Error(err))
	} else {
		zap.L().Info("HTTP server shutdown complete")
	}
}
