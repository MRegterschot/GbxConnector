package listeners

import (
	"github.com/MRegterschot/GbxConnector/config"
	"github.com/MRegterschot/GbxConnector/sockets"
	"github.com/MRegterschot/GbxRemoteGo/gbxclient"
)

func OnConnect(client *gbxclient.GbxClient, serverId int) {
	onConnectChan := make(chan any, 1)
	client.Events.On("connect", onConnectChan)

	go func() {
		for range onConnectChan {
			config.AppEnv.Servers[serverId].IsConnected = true
			sockets.BroadcastServers(config.AppEnv.Servers.ToServerResponses())
			return
		}
	}()
}

func OnDisconnect(client *gbxclient.GbxClient, serverId int) {
	onDisconnectChan := make(chan any, 1)
	client.Events.On("disconnect", onDisconnectChan)

	go func() {
		for range onDisconnectChan {
			config.AppEnv.Servers[serverId].IsConnected = false
			sockets.BroadcastServers(config.AppEnv.Servers.ToServerResponses())
			return
		}
	}()
}
