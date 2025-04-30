package listeners

import (
	"github.com/MRegterschot/GbxConnector/config"
	"github.com/MRegterschot/GbxConnector/handlers"
	"github.com/MRegterschot/GbxConnector/structs"
)

func OnConnect(server *structs.Server) {
	onConnectChan := make(chan any, 1)
	server.Client.Events.On("connect", onConnectChan)

	go func() {
		for range onConnectChan {
			handlers.BroadcastServers(config.AppEnv.Servers.ToServerResponses())
		}
	}()
}

func OnDisconnect(server *structs.Server) {
	onDisconnectChan := make(chan any, 1)
	server.Client.Events.On("disconnect", onDisconnectChan)

	go func() {
		for range onDisconnectChan {
			handlers.BroadcastServers(config.AppEnv.Servers.ToServerResponses())
		}
	}()
}
