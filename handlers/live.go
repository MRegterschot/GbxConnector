package handlers

import (
	"net/http"

	"github.com/MRegterschot/GbxConnector/config"
	"github.com/MRegterschot/GbxConnector/structs"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var liveSockets = make(map[string]*structs.SocketClients)

func GetLiveSocket(serverUuid string) *structs.SocketClients {
	if _, ok := liveSockets[serverUuid]; !ok {
		liveSockets[serverUuid] = &structs.SocketClients{
			Clients: make(map[*websocket.Conn]bool),
		}
	}
	return liveSockets[serverUuid]
}

func HandleLiveConnection(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverUuid := vars["uuid"]

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		zap.L().Error("Failed to upgrade connection", zap.Error(err))
		return
	}

	// Save connection
	ls := GetLiveSocket(serverUuid)
	ls.ClientsMu.Lock()
	ls.Clients[conn] = true
	ls.ClientsMu.Unlock()

	// Get current live info of the server
	var liveInfo *structs.LiveInfo
	for _, server := range config.AppEnv.Servers {
		if server.Uuid == serverUuid {
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
				zap.L().Info("WebSocket connection closed", zap.String("remoteAddr", conn.RemoteAddr().String()), zap.String("server_uuid", serverUuid))
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
func BroadcastLive(serverUuid string, message any) {
	ls := GetLiveSocket(serverUuid)
	if ls == nil {
		zap.L().Error("Live socket not found", zap.String("server_uuid", serverUuid))
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
