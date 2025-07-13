package handlers

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/MRegterschot/GbxConnector/config"
	"github.com/MRegterschot/GbxConnector/lib"
	"github.com/MRegterschot/GbxConnector/structs"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// Define WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for WebSocket connections
		return true
	},
}

type ServerSocket struct {
	Clients       map[*websocket.Conn]bool // Connected clients
	ClientsMu     sync.Mutex
	Subscriptions map[*websocket.Conn]map[string]bool // conn => set of serverUuids
}

var serverSocket = &ServerSocket{
	Clients:       make(map[*websocket.Conn]bool),
	Subscriptions: make(map[*websocket.Conn]map[string]bool),
}

type ServerAdderFunc func(server *structs.Server) (*structs.Server, error)
type ServerRemoverFunc func(serverUuid string) error
type ServerUpdaterFunc func(serverUuid string, server *structs.Server) (*structs.Server, error)

var addServerFunc ServerAdderFunc
var removeServerFunc ServerRemoverFunc
var updateServerFunc ServerUpdaterFunc

func SetAddServerFunc(fn ServerAdderFunc) {
	addServerFunc = fn
}

func SetRemoveServerFunc(fn ServerRemoverFunc) {
	removeServerFunc = fn
}

func SetUpdateServerFunc(fn ServerUpdaterFunc) {
	updateServerFunc = fn
}

// WebSocket connection handler
func HandleServersConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		zap.L().Error("Failed to upgrade connection", zap.Error(err))
		return
	}

	// Parse query parameters for serverUuids
	query := r.URL.Query()
	serverUuids := query["serverUuid"] // expecting ?serverUuid=uuid1&serverUuid=uuid2

	subscriptionSet := make(map[string]bool)
	for _, uuid := range serverUuids {
		subscriptionSet[uuid] = true
	}

	// Save connection and subscriptions
	serverSocket.ClientsMu.Lock()
	serverSocket.Clients[conn] = true
	serverSocket.Subscriptions[conn] = subscriptionSet
	serverSocket.ClientsMu.Unlock()

	zap.L().Info("New WebSocket connection established", zap.String("remoteAddr", conn.RemoteAddr().String()), zap.Strings("subscriptions", serverUuids))

	// Send initial message for subscribed servers only
	allServers := config.AppEnv.Servers.ToServerResponses()
	filteredServers := lib.FilterServersByUuid(allServers, subscriptionSet)

	if err := conn.WriteJSON(filteredServers); err != nil {
		zap.L().Error("Failed to send initial message to client", zap.Error(err))
		serverSocket.ClientsMu.Lock()
		delete(serverSocket.Clients, conn)
		delete(serverSocket.Subscriptions, conn)
		serverSocket.ClientsMu.Unlock()
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
				delete(serverSocket.Subscriptions, conn)
				serverSocket.ClientsMu.Unlock()
				conn.Close()
				break
			}
		}
	}()
}

// Broadcast message to all connected clients
// This function is generic and can be used to send any type of message
func BroadcastServers(allServers []structs.ServerResponse) {
	serverSocket.ClientsMu.Lock()
	defer serverSocket.ClientsMu.Unlock()

	for conn := range serverSocket.Clients {
		subscriptionSet := serverSocket.Subscriptions[conn]

		filteredServers := lib.FilterServersByUuid(allServers, subscriptionSet)

		if err := conn.WriteJSON(filteredServers); err != nil {
			zap.L().Error("Failed to send message to client", zap.Error(err))
			conn.Close()
			delete(serverSocket.Clients, conn)
			delete(serverSocket.Subscriptions, conn)
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

	newServer, err := addServerFunc(&server)
	server.ResetLiveInfo()

	if err != nil {
		zap.L().Error("Failed to add server", zap.Error(err))
		http.Error(w, "Failed to add server", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(newServer.ToServerResponse()); err != nil {
		zap.L().Error("Failed to encode server response", zap.Error(err))
		http.Error(w, "Failed to encode server response", http.StatusInternalServerError)
		return
	}
}

// HandleDeleteServer handles requests to delete a server
func HandleDeleteServer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverUuid := vars["uuid"]

	if removeServerFunc == nil {
		zap.L().Error("Remove server function not set")
		http.Error(w, "Server configuration error", http.StatusInternalServerError)
		return
	}

	if err := removeServerFunc(serverUuid); err != nil {
		zap.L().Error("Failed to remove server", zap.Error(err))
		http.Error(w, "Failed to remove server", http.StatusInternalServerError)
		return
	}

	// Broadcast updated server list
	servers := config.AppEnv.Servers.ToServerResponses()
	BroadcastServers(servers)
	w.WriteHeader(http.StatusOK)
}

// HandleUpdateServer handles requests to update a server
func HandleUpdateServer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverUuid := vars["uuid"]

	var server structs.Server
	if err := json.NewDecoder(r.Body).Decode(&server); err != nil {
		zap.L().Error("Failed to decode server", zap.Error(err))
		http.Error(w, "Failed to decode server", http.StatusBadRequest)
		return
	}

	if updateServerFunc == nil {
		zap.L().Error("Update server function not set")
		http.Error(w, "Server configuration error", http.StatusInternalServerError)
		return
	}

	updatedServer, err := updateServerFunc(serverUuid, &server)
	if err != nil {
		zap.L().Error("Failed to update server", zap.Error(err))
		http.Error(w, "Failed to update server", http.StatusInternalServerError)
		return
	}

	server.ResetLiveInfo()

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(updatedServer.ToServerResponse()); err != nil {
		zap.L().Error("Failed to encode server response", zap.Error(err))
		http.Error(w, "Failed to encode server response", http.StatusInternalServerError)
		return
	}
}
