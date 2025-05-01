package handlers

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/MRegterschot/GbxConnector/config"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type ListenerSocket struct {
	Clients   map[*websocket.Conn]bool // Connected clients
	ClientsMu sync.Mutex
}

var listenerSockets = make(map[int]*ListenerSocket) // Map of listener sockets by server ID

func GetListenerSocket(serverId int) *ListenerSocket {
	if _, ok := listenerSockets[serverId]; !ok {
		listenerSockets[serverId] = &ListenerSocket{
			Clients: make(map[*websocket.Conn]bool),
		}
	}
	return listenerSockets[serverId]
}

// WebSocket connection handler
func HandleListenerConnection(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverIDStr := vars["serverId"]

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
	ls := GetListenerSocket(serverId)
	ls.ClientsMu.Lock()
	ls.Clients[conn] = true
	ls.ClientsMu.Unlock()

	// Get current active map of the server
	var activeMap string
	for _, server := range config.AppEnv.Servers {
		if server.Id == serverId {
			activeMap = server.ActiveMap
			break
		}
	}

	// Send initial message to the client
	if err := conn.WriteJSON(activeMap); err != nil {
		zap.L().Error("Failed to send initial message to client", zap.Error(err))
		conn.Close()
		return
	}

	// Handle disconnection
	go func() {
		for {
			if _, _, err := conn.NextReader(); err != nil {
				zap.L().Info("WebSocket connection closed", zap.String("remoteAddr", conn.RemoteAddr().String()))
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
func BroadcastListener(serverId int, message any) {
	ls := GetListenerSocket(serverId)
	if ls == nil {
		zap.L().Error("Listener socket not found", zap.Int("serverId", serverId))
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
