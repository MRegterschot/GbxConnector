package structs

import (
	"sync"

	"github.com/gorilla/websocket"
)

type SocketClients struct {
	Clients   map[*websocket.Conn]bool // Connected clients
	ClientsMu sync.Mutex
}