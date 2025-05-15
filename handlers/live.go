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

var liveSockets = make(map[int]*structs.SocketClients)

func GetLiveSocket(serverId int) *structs.SocketClients {
	if _, ok := liveSockets[serverId]; !ok {
		liveSockets[serverId] = &structs.SocketClients{
			Clients: make(map[*websocket.Conn]bool),
		}
	}
	return liveSockets[serverId]
}

func HandleLiveConnection(w http.ResponseWriter, r *http.Request) {
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
	ls := GetLiveSocket(serverId)
	ls.ClientsMu.Lock()
	ls.Clients[conn] = true
	ls.ClientsMu.Unlock()

	// Get current live info of the server
	var liveInfo *structs.LiveInfo
	for _, server := range config.AppEnv.Servers {
		if server.Id == serverId {
			liveInfo = server.Info.LiveInfo
			break
		}
	}

	// Send initial message to the client
	if err := conn.WriteJSON(map[string]*structs.LiveInfo{
		"beginMatch": liveInfo,
	}); err != nil {
		zap.L().Error("Failed to send initial message to client", zap.Error(err))
		ls.ClientsMu.Lock()
		delete(ls.Clients, conn)
		ls.ClientsMu.Unlock()
		conn.Close()
		return
	}

	// Handle disconnection
	go func() {
		for {
			if _, _, err := conn.NextReader(); err != nil {
				zap.L().Info("WebSocket connection closed", zap.String("remoteAddr", conn.RemoteAddr().String()), zap.Int("serverId", serverId))
				ls.ClientsMu.Lock()
				delete(ls.Clients, conn)
				ls.ClientsMu.Unlock()
				conn.Close()
				break
			}
		}
	}()
}

// Broadcast message to all connected clients
func BroadcastLive(serverId int, message any) {
	ls := GetLiveSocket(serverId)
	if ls == nil {
		zap.L().Error("Live socket not found", zap.Int("serverId", serverId))
		return
	}

	ls.ClientsMu.Lock()
	defer ls.ClientsMu.Unlock()

	for conn := range ls.Clients {
		if err := conn.WriteJSON(message); err != nil {
			zap.L().Error("Failed to send message to client", zap.Error(err))
			conn.Close()
			delete(ls.Clients, conn)
		}
	}
}
