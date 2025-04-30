package sockets

import (
	"net/http"
	"slices"
	"sync"

	"github.com/MRegterschot/GbxConnector/config"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// Define WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		allowedOrigins := config.AppEnv.CorsOrigins

		// If no origin header is present, deny the request
		if origin == "" {
			zap.L().Debug("WebSocket request missing Origin header")
			return false
		}

		// Check for empty allowed origins
		if len(allowedOrigins) == 0 {
			return true
		}

		// Check for allow all origins
		if slices.Contains(allowedOrigins, "*") {
			return true
		}

		// Check if the origin is in the allowed origins
		return slices.Contains(allowedOrigins, origin)
	},
}

var (
	clients   = make(map[*websocket.Conn]bool) // Connected clients
	clientsMu sync.Mutex
)

// WebSocket connection handler
func HandleServersConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		zap.L().Error("Failed to upgrade connection", zap.Error(err))
		return
	}

	// Save connection
	clientsMu.Lock()
	clients[conn] = true
	clientsMu.Unlock()

	zap.L().Info("New WebSocket connection established", zap.String("remoteAddr", conn.RemoteAddr().String()))

	// Send initial message to the client
	if err := conn.WriteJSON(config.AppEnv.Servers.ToServerResponses()); err != nil {
		zap.L().Error("Failed to send initial message to client", zap.Error(err))
		conn.Close()
		return
	}

	// Handle disconnection
	go func() {
		for {
			if _, _, err := conn.NextReader(); err != nil {
				zap.L().Info("WebSocket connection closed", zap.String("remoteAddr", conn.RemoteAddr().String()))
				clientsMu.Lock()
				delete(clients, conn)
				clientsMu.Unlock()
				conn.Close()
				break
			}
		}
	}()
}

// Broadcast message to all connected clients
// This function is generic and can be used to send any type of message
func BroadcastServers[T any](msg T) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	for conn := range clients {
		if err := conn.WriteJSON(msg); err != nil {
			zap.L().Error("Failed to send message to client", zap.Error(err))
			conn.Close()
			delete(clients, conn)
		}
	}
}
