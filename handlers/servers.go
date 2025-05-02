package handlers

import (
	"encoding/json"
	"net/http"
	"slices"
	"sync"

	"github.com/MRegterschot/GbxConnector/config"
	"github.com/MRegterschot/GbxConnector/structs"
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

type ServerSocket struct {
	Clients map[*websocket.Conn]bool // Connected clients
	ClientsMu sync.Mutex
}

var serverSocket = &ServerSocket{
	Clients: make(map[*websocket.Conn]bool),
}

// WebSocket connection handler
func HandleServersConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		zap.L().Error("Failed to upgrade connection", zap.Error(err))
		return
	}

	// Save connection
	serverSocket.ClientsMu.Lock()
	serverSocket.Clients[conn] = true
	serverSocket.ClientsMu.Unlock()

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
				serverSocket.ClientsMu.Lock()
				delete(serverSocket.Clients, conn)
				serverSocket.ClientsMu.Unlock()
				conn.Close()
				break
			}
		}
	}()
}

// Broadcast message to all connected clients
// This function is generic and can be used to send any type of message
func BroadcastServers[T any](msg T) {
	serverSocket.ClientsMu.Lock()
	defer serverSocket.ClientsMu.Unlock()
	for conn := range serverSocket.Clients {
		if err := conn.WriteJSON(msg); err != nil {
			zap.L().Error("Failed to send message to client", zap.Error(err))
			conn.Close()
			delete(serverSocket.Clients, conn)
		}
	}
}

// Handle GET request to retrieve server information
func HandleGetServers(w http.ResponseWriter, r *http.Request) {
	servers := config.AppEnv.Servers.ToServerResponses()
	if err := json.NewEncoder(w).Encode(servers); err != nil {
		zap.L().Error("Failed to encode servers response", zap.Error(err))
		http.Error(w, "Failed to encode servers response", http.StatusInternalServerError)
		return
	}
}

func HandleAddServer(w http.ResponseWriter, r *http.Request) {
	var server structs.Server
	if err := json.NewDecoder(r.Body).Decode(&server); err != nil {
		zap.L().Error("Failed to decode server", zap.Error(err))
		http.Error(w, "Failed to decode server", http.StatusBadRequest)
		return
	}

	config.AppEnv.Servers = append(config.AppEnv.Servers, &server)
	BroadcastServers(config.AppEnv.Servers.ToServerResponses())
	w.WriteHeader(http.StatusCreated)
}