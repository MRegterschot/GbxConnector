package handlers

import (
	"net/http"
	"strconv"

	"github.com/MRegterschot/GbxConnector/config"
	"github.com/MRegterschot/GbxConnector/structs"
	gbxstructs "github.com/MRegterschot/GbxRemoteGo/structs"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var playersSockets = make(map[int]*structs.SocketClients) // Map of socket clients by server ID

func GetPlayersSocket(serverId int) *structs.SocketClients {
	if _, ok := playersSockets[serverId]; !ok {
		playersSockets[serverId] = &structs.SocketClients{
			Clients: make(map[*websocket.Conn]bool),
		}
	}
	return playersSockets[serverId]
}

// WebSocket connection handler
func HandlePlayersConnection(w http.ResponseWriter, r *http.Request) {
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
	ps := GetPlayersSocket(serverId)
	ps.ClientsMu.Lock()
	ps.Clients[conn] = true
	ps.ClientsMu.Unlock()

	// Get current active map of the server
	activePlayers := make([]gbxstructs.TMPlayerInfo, 0)
	for _, server := range config.AppEnv.Servers {
		if server.Id == serverId {
			if server.ActivePlayers != nil {
				activePlayers = server.ActivePlayers
			}
			break
		}
	}

	// Send initial message to the client
	if err := conn.WriteJSON(map[string][]gbxstructs.TMPlayerInfo{
		"playerList": activePlayers,
	}); err != nil {
		zap.L().Error("Failed to send initial message to client", zap.Error(err))
		conn.Close()
		return
	}

	// Handle disconnection
	go func() {
		for {
			if _, _, err := conn.NextReader(); err != nil {
				zap.L().Info("WebSocket connection closed", zap.String("remoteAddr", conn.RemoteAddr().String()), zap.Int("serverId", serverId))
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
func BroadcastPlayers(serverId int, message any) {
	ps := GetPlayersSocket(serverId)
	if ps == nil {
		zap.L().Error("Players socket not found", zap.Int("serverId", serverId))
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
