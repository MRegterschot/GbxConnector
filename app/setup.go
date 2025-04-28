package app

import (
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
