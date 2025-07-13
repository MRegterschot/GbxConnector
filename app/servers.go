package app

import (
	"context"
	"errors"

	"slices"

	"github.com/MRegterschot/GbxConnector/config"
	"github.com/MRegterschot/GbxConnector/handlers"
	"github.com/MRegterschot/GbxConnector/lib"
	"github.com/MRegterschot/GbxConnector/structs"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func ShutdownServers(servers []*structs.Server) {
	for _, server := range servers {
		if server.CancelFunc != nil {
			zap.L().Info("Shutting down server", zap.String("host", server.Host), zap.Int("port", server.XMLRPCPort))
			server.CancelFunc()
		}
	}
}

func ShutdownServer(server *structs.Server) {
	if server.CancelFunc != nil {
		zap.L().Info("Shutting down server", zap.String("host", server.Host), zap.Int("port", server.XMLRPCPort))
		server.CancelFunc()
	}
}

// AddServer adds a new server to the configuration and sets it up
func AddServer(server *structs.Server) (*structs.Server, error) {
	for _, s := range config.AppEnv.Servers {
		if s.Host == server.Host && s.XMLRPCPort == server.XMLRPCPort {
			zap.L().Error("Server already exists", zap.String("host", server.Host), zap.Int("port", server.XMLRPCPort))
			return nil, errors.New("server already exists")
		}
	}

	server.Uuid = uuid.NewString()

	config.AppEnv.Servers = append(config.AppEnv.Servers, server)
	if err := lib.WriteFile("./servers.json", &config.AppEnv.Servers); err != nil {
		zap.L().Error("Failed to write servers.json", zap.Error(err))
		return nil, err
	}

	zap.L().Info("New server added", zap.String("server_uuid", server.Uuid))

	GetClient(server)
	handlers.GetMapSocket(server.Uuid)
	handlers.GetPlayersSocket(server.Uuid)

	ctx, cancel := context.WithCancel(context.Background())
	server.Ctx = ctx
	server.CancelFunc = cancel

	go StartReconnectLoop(ctx, server)

	handlers.BroadcastServers(config.AppEnv.Servers.ToServerResponses())

	return server, nil
}

func DeleteServer(serverUuid string) error {
	for i, server := range config.AppEnv.Servers {
		if server.Uuid == serverUuid {
			config.AppEnv.Servers = slices.Delete(config.AppEnv.Servers, i, i+1)
			if err := lib.WriteFile("./servers.json", &config.AppEnv.Servers); err != nil {
				zap.L().Error("Failed to write servers.json", zap.Error(err))
				return err
			}
			zap.L().Info("Server deleted", zap.String("server_uuid", serverUuid))
			handlers.BroadcastServers(config.AppEnv.Servers.ToServerResponses())
			ShutdownServer(server)
			return nil
		}
	}
	return errors.New("server not found")
}

func UpdateServer(serverUuid string, serverInput *structs.Server) (*structs.Server, error) {
	for _, server := range config.AppEnv.Servers {
		if server.Uuid == serverUuid {
			server.UpdateServer(
				serverInput.Name,
				serverInput.Description,
				serverInput.Host,
				serverInput.XMLRPCPort,
				serverInput.User,
				serverInput.Pass,
				serverInput.FMUrl,
			)

			if err := lib.WriteFile("./servers.json", &config.AppEnv.Servers); err != nil {
				zap.L().Error("Failed to write servers.json", zap.Error(err))
				return nil, err
			}
			zap.L().Info("Server updated", zap.String("server_uuid", serverUuid))

			ShutdownServer(server)
			GetClient(server)
			handlers.GetMapSocket(server.Uuid)
			handlers.GetPlayersSocket(server.Uuid)

			ctx, cancel := context.WithCancel(context.Background())
			server.Ctx = ctx
			server.CancelFunc = cancel
			go StartReconnectLoop(ctx, server)

			handlers.BroadcastServers(config.AppEnv.Servers.ToServerResponses())
			return server, nil
		}
	}
	return nil, errors.New("server not found")
}
