package handlers

import (
	"net/http"
	"strconv"

	"github.com/MRegterschot/GbxConnector/config"
	"github.com/MRegterschot/GbxConnector/structs"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var mapSockets = make(map[int]*structs.SocketClients) // Map of socket clients by server ID

func GetMapSocket(serverId int) *structs.SocketClients {
	if _, ok := mapSockets[serverId]; !ok {
		mapSockets[serverId] = &structs.SocketClients{
			Clients: make(map[*websocket.Conn]bool),
		}
	}
	return mapSockets[serverId]
}

// WebSocket connection handler
func HandleMapConnection(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverIDStr := vars["id"]

	serverId, err := strconv.Atoi(serverIDStr)
	if err != nil {
		http.Error(w, "Invalid server ID", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		zap.L().Error("Failed to upgrade connection", zap.Error(err))
		return
	}

	// Save connection
	ms := GetMapSocket(serverId)
	ms.ClientsMu.Lock()
	ms.Clients[conn] = true
	ms.ClientsMu.Unlock()

	// Get current active map of the server
	var activeMap string
	for _, server := range config.AppEnv.Servers {
		if server.Id == serverId {
			activeMap = server.Info.ActiveMap
			break
		}
	}

	// Send initial message to the client
	if err := conn.WriteJSON(activeMap); err != nil {
		zap.L().Error("Failed to send initial message to client", zap.Error(err))
		ms.ClientsMu.Lock()
		delete(ms.Clients, conn)
		ms.ClientsMu.Unlock()
		conn.Close()
		return
	}

	// Handle disconnection
	go func() {
		for {
			if _, _, err := conn.NextReader(); err != nil {
				zap.L().Info("WebSocket connection closed", zap.String("remoteAddr", conn.RemoteAddr().String()), zap.Int("serverId", serverId))
				ms.ClientsMu.Lock()
				delete(ms.Clients, conn)
				ms.ClientsMu.Unlock()
				conn.Close()
				break
			}
		}
	}()
}

// Broadcast message to all connected clients
func BroadcastMap(serverId int, message any) {
	ms := GetMapSocket(serverId)
	if ms == nil {
		zap.L().Error("Map socket not found", zap.Int("serverId", serverId))
		return
	}

	ms.ClientsMu.Lock()
	defer ms.ClientsMu.Unlock()

	for conn := range ms.Clients {
		if err := conn.WriteJSON(message); err != nil {
			zap.L().Error("Failed to send message to client", zap.Error(err))
			conn.Close()
			delete(ms.Clients, conn)
		}
	}
}
