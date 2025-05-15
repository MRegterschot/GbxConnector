package app

import (
	"context"
	"errors"
	"time"

	"github.com/MRegterschot/GbxConnector/config"
	"github.com/MRegterschot/GbxConnector/listeners"
	"github.com/MRegterschot/GbxConnector/structs"
	"github.com/MRegterschot/GbxRemoteGo/gbxclient"
	"go.uber.org/zap"
)

func GetClient(server *structs.Server) error {
	if server.Client != nil {
		return nil
	}

	server.Client = gbxclient.NewGbxClient(server.Host, server.XMLRPCPort, gbxclient.Options{})

	// Add listeners
	listeners.AddConnectionListeners(server)
	listeners.AddMapListeners(server)
	pl := listeners.AddPlayersListeners(server)
	listeners.AddLiveListeners(server)

	if err := ConnectClient(server); err != nil {
		return err
	}

	pl.SyncPlayerList()

	return nil
}

func ConnectClient(server *structs.Server) error {
	if server.Client == nil {
		zap.L().Error("Client is nil")
		return errors.New("client is nil")
	}

	zap.L().Debug("Connecting to server", zap.String("host", server.Host), zap.Int("port", server.XMLRPCPort))
	if err := server.Client.Connect(); err != nil {
		zap.L().Debug("Failed to connect to server", zap.Error(err))
		return err
	}

	zap.L().Info("Authenticating with server", zap.Int("id", server.Id))
	if err := server.Client.Authenticate(server.User, server.Pass); err != nil {
		zap.L().Error("Failed to authenticate with server", zap.Error(err))
		return err
	}

	zap.L().Info("Connected to server", zap.String("host", server.Host), zap.Int("port", server.XMLRPCPort))

	server.Client.EnableCallbacks(true)
	server.Client.SetApiVersion("2023-04-16")
	server.Client.TriggerModeScriptEventArray("XmlRpc.EnableCallbacks", []string{"true"})

	// Set the active map to the current map
	mapInfo, err := server.Client.GetCurrentMapInfo()
	if err != nil {
		zap.L().Error("Failed to get current map info", zap.Error(err))
	}

	server.ActiveMap = mapInfo.UId

	return nil
}

func StartReconnectLoop(ctx context.Context, server *structs.Server) {
	go func() {
		ticker := time.NewTicker(config.AppEnv.ReconnectInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				zap.L().Info("Reconnect loop stopped", zap.Int("server_id", server.Id))
				return

			case <-ticker.C:
				if server.Client == nil || !server.Client.IsConnected {
					zap.L().Debug("Client disconnected or missing, attempting reconnect", zap.Int("server_id", server.Id))

					if server.Client == nil {
						if err := GetClient(server); err != nil {
							zap.L().Debug("Failed to get client", zap.Error(err))
							continue
						}
					} else {
						if err := ConnectClient(server); err != nil {
							zap.L().Debug("Failed to reconnect to server", zap.Error(err))
							continue
						}
					}
				}
			}
		}
	}()
}
