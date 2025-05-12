package handlers

import (
	"net/http"
	"strconv"

	"github.com/MRegterschot/GbxConnector/structs"
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

	// Send initial message to the client
	if err := conn.WriteJSON("Connected to players socket"); err != nil {
		zap.L().Error("Failed to send initial message to client", zap.Error(err))
		conn.Close()
		return
	}
}