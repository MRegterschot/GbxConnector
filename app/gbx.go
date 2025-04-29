package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/MRegterschot/GbxConnector/listeners"
	"github.com/MRegterschot/GbxConnector/structs"
	"github.com/MRegterschot/GbxRemoteGo/gbxclient"
	"go.uber.org/zap"
)

var clients = make(map[int]*gbxclient.GbxClient)

func GetClient(server structs.Server) (*gbxclient.GbxClient, error) {
	if client, ok := clients[server.Id]; ok {
		return client, nil
	}

	client := gbxclient.NewGbxClient(server.Host, server.XMLRPCPort, gbxclient.Options{})

	listeners.OnConnect(client, server.Id)
	listeners.OnDisconnect(client, server.Id)

	if err := ConnectClient(client, server); err != nil {
		zap.L().Error("Failed to connect to server", zap.Error(err))
		return nil, err
	}

	clients[server.Id] = client
	return client, nil
}

func ConnectClient(client *gbxclient.GbxClient, server structs.Server) error {
	if client == nil {
		zap.L().Error("Client is nil")
		return errors.New("client is nil")
	}

	zap.L().Info("Connecting to server", zap.String("host", server.Host), zap.Int("port", server.XMLRPCPort))
	if err := client.Connect(); err != nil {
		zap.L().Error("Failed to connect to server", zap.Error(err))
		return err
	}

	zap.L().Info("Authenticating with server", zap.String("user", server.User))
	if err := client.Authenticate(server.User, server.Pass); err != nil {
		zap.L().Error("Failed to authenticate with server", zap.Error(err))
		return err
	}

	zap.L().Info("Connected to server", zap.String("host", server.Host), zap.Int("port", server.XMLRPCPort))
	return nil
}

func StartReconnectLoop(ctx context.Context, server structs.Server) {
	go func() {
		ticker := time.NewTicker(6 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				zap.L().Info("Reconnect loop stopped", zap.Int("server_id", server.Id))
				return

			case <-ticker.C:
				fmt.Println("Checking connection status...")
				client, exists := clients[server.Id]

				if !exists || client == nil || !client.IsConnected {
					zap.L().Warn("Client disconnected or missing, attempting reconnect", zap.Int("server_id", server.Id))

					if client == nil || !exists {
						var err error
						clients[server.Id], err = GetClient(server)
						if err != nil {
							zap.L().Error("Failed to get client", zap.Error(err))
							continue
						}
					} else {
						if err := ConnectClient(client, server); err != nil {
							zap.L().Error("Failed to reconnect to server", zap.Error(err))
							continue
						}
					}
				}
			}
		}
	}()
}