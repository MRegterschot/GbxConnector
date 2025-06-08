package handlers

import (
	"net/http"

	"github.com/MRegterschot/GbxConnector/config"
	"github.com/MRegterschot/GbxConnector/structs"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var playersSockets = make(map[string]*structs.SocketClients) // Map of socket clients by server ID

func GetPlayersSocket(serverUuid string) *structs.SocketClients {
	if _, ok := playersSockets[serverUuid]; !ok {
		playersSockets[serverUuid] = &structs.SocketClients{
			Clients: make(map[*websocket.Conn]bool),
		}
	}
	return playersSockets[serverUuid]
}

// WebSocket connection handler
func HandlePlayersConnection(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverUuid := vars["uuid"]

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		zap.L().Error("Failed to upgrade connection", zap.Error(err))
		return
	}

	// Save connection
	ps := GetPlayersSocket(serverUuid)
	ps.ClientsMu.Lock()
	ps.Clients[conn] = true
	ps.ClientsMu.Unlock()

	// Get current active map of the server
	activePlayers := make([]structs.PlayerInfo, 0)
	for _, server := range config.AppEnv.Servers {
		if server.Uuid == serverUuid {
			if server.Info.ActivePlayers != nil {
				activePlayers = server.Info.ActivePlayers
			}
			break
		}
	}

	// Send initial message to the client
	if err := conn.WriteJSON(map[string][]structs.PlayerInfo{
		"playerList": activePlayers,
	}); err != nil {
		zap.L().Error("Failed to send initial message to client", zap.Error(err))
		ps.ClientsMu.Lock()
		delete(ps.Clients, conn)
		ps.ClientsMu.Unlock()
		conn.Close()
		return
	}

	// Handle disconnection
	go func() {
		for {
			if _, _, err := conn.NextReader(); err != nil {
				zap.L().Info("WebSocket connection closed", zap.String("remoteAddr", conn.RemoteAddr().String()), zap.String("server_uuid", serverUuid))
				ps.ClientsMu.Lock()
				delete(ps.Clients, conn)
				ps.ClientsMu.Unlock()
				conn.Close()
				break
			}
		}
	}()
}

// Broadcast message to all connected clients
func BroadcastPlayers(serverUuid string, message any) {
	ps := GetPlayersSocket(serverUuid)
	if ps == nil {
		zap.L().Error("Players socket not found", zap.String("server_uuid", serverUuid))
		return
	}

	ps.ClientsMu.Lock()
	defer ps.ClientsMu.Unlock()

	for conn := range ps.Clients {
		if err := conn.WriteJSON(message); err != nil {
			zap.L().Error("Failed to send message to client", zap.Error(err))
			conn.Close()
			delete(ps.Clients, conn)
		}
	}
}
