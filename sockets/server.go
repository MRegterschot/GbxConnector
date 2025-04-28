package sockets

import (
	"net/http"

	"github.com/MRegterschot/GbxConnector/config"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// Define WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		allowedOrigin := config.AppEnv.CorsOrigin

		if origin == "" {
			zap.L().Warn("WebSocket request missing Origin header")
			return false
		}

		if origin != allowedOrigin {
			zap.L().Warn("WebSocket Origin not allowed", zap.String("origin", origin), zap.String("allowed", allowedOrigin))
			return false
		}

		return true
	},
}

// WebSocket connection handler
func HandleServersConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		zap.L().Error("Failed to upgrade connection", zap.Error(err))
		return
	}
	defer conn.Close()

	zap.L().Info("WebSocket connection established", zap.String("remote_addr", r.RemoteAddr))

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			zap.L().Error("Error reading message", zap.Error(err))
			return
		}

		zap.L().Info("Received message", zap.String("message", string(p)))

		if err := conn.WriteMessage(messageType, p); err != nil {
			zap.L().Error("Error writing message", zap.Error(err))
			return
		}
	}
}
