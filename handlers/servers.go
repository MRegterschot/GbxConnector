package handlers

import (
	"encoding/json"
	"net/http"
	"slices"
	"strconv"
	"sync"

	"github.com/MRegterschot/GbxConnector/config"
	"github.com/MRegterschot/GbxConnector/structs"
	"github.com/gorilla/mux"
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
	Clients   map[*websocket.Conn]bool // Connected clients
	ClientsMu sync.Mutex
}

var serverSocket = &ServerSocket{
	Clients: make(map[*websocket.Conn]bool),
}

type ServerAdderFunc func(server *structs.Server) error
type ServerRemoverFunc func(serverId int) error

var addServerFunc ServerAdderFunc
var removeServerFunc ServerRemoverFunc

func SetAddServerFunc(fn ServerAdderFunc) {
	addServerFunc = fn
}

func SetRemoveServerFunc(fn ServerRemoverFunc) {
	removeServerFunc = fn
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

// HandleAddServer handles requests to add a new server
func HandleAddServer(w http.ResponseWriter, r *http.Request) {
	var server structs.Server
	if err := json.NewDecoder(r.Body).Decode(&server); err != nil {
		zap.L().Error("Failed to decode server", zap.Error(err))
		http.Error(w, "Failed to decode server", http.StatusBadRequest)
		return
	}

	if addServerFunc == nil {
		zap.L().Error("Add server function not set")
		http.Error(w, "Server configuration error", http.StatusInternalServerError)
		return
	}

	if err := addServerFunc(&server); err != nil {
		zap.L().Error("Failed to add server", zap.Error(err))
		http.Error(w, "Failed to add server", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// HandleDeleteServer handles requests to delete a server
func HandleDeleteServer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverIDStr := vars["id"]

	serverId, err := strconv.Atoi(serverIDStr)
	if err != nil {
		zap.L().Error("Invalid server ID", zap.Error(err))
		http.Error(w, "Invalid server ID", http.StatusBadRequest)
		return
	}

	if removeServerFunc == nil {
		zap.L().Error("Remove server function not set")
		http.Error(w, "Server configuration error", http.StatusInternalServerError)
		return
	}

	if err := removeServerFunc(serverId); err != nil {
		zap.L().Error("Failed to remove server", zap.Error(err))
		http.Error(w, "Failed to remove server", http.StatusInternalServerError)
		return
	}

	// Broadcast updated server list
	servers := config.AppEnv.Servers.ToServerResponses()
	BroadcastServers(servers)
	w.WriteHeader(http.StatusOK)
}
