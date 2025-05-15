package handlers

import (
	"net/http"
	"strconv"

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

func BroadcastLive(serverId int, message map[string]any) {
	ls := GetLiveSocket(serverId)
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
