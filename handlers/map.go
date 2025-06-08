package handlers

import (
	"net/http"

	"github.com/MRegterschot/GbxConnector/config"
	"github.com/MRegterschot/GbxConnector/structs"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var mapSockets = make(map[string]*structs.SocketClients) // Map of socket clients by server ID

func GetMapSocket(serverUuid string) *structs.SocketClients {
	if _, ok := mapSockets[serverUuid]; !ok {
		mapSockets[serverUuid] = &structs.SocketClients{
			Clients: make(map[*websocket.Conn]bool),
		}
	}
	return mapSockets[serverUuid]
}

// WebSocket connection handler
func HandleMapConnection(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverUuid := vars["uuid"]

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		zap.L().Error("Failed to upgrade connection", zap.Error(err))
		return
	}

	// Save connection
	ms := GetMapSocket(serverUuid)
	ms.ClientsMu.Lock()
	ms.Clients[conn] = true
	ms.ClientsMu.Unlock()

	// Get current active map of the server
	var activeMap string
	for _, server := range config.AppEnv.Servers {
		if server.Uuid == serverUuid {
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
				zap.L().Info("WebSocket connection closed", zap.String("remoteAddr", conn.RemoteAddr().String()), zap.String("server_uuid", serverUuid))
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
func BroadcastMap(serverUuid string, message any) {
	ms := GetMapSocket(serverUuid)
	if ms == nil {
		zap.L().Error("Map socket not found", zap.String("server_uuid", serverUuid))
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
