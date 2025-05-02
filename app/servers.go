package app

import (
	"context"
	"errors"

	"slices"

	"github.com/MRegterschot/GbxConnector/config"
	"github.com/MRegterschot/GbxConnector/handlers"
	"github.com/MRegterschot/GbxConnector/lib"
	"github.com/MRegterschot/GbxConnector/structs"
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
	if server.Id == 0 {
		maxID := 0
		for _, s := range config.AppEnv.Servers {
			if s.Host == server.Host && s.XMLRPCPort == server.XMLRPCPort {
				zap.L().Error("Server already exists", zap.String("host", server.Host), zap.Int("port", server.XMLRPCPort))
				return nil, errors.New("server already exists")
			}

			if s.Id > maxID {
				maxID = s.Id
			}
		}
		server.Id = maxID + 1
	}

	server.IsLocal = lib.IsLocalHostname(server.Host)

	config.AppEnv.Servers = append(config.AppEnv.Servers, server)
	if err := lib.WriteFile("./servers.json", &config.AppEnv.Servers); err != nil {
		zap.L().Error("Failed to write servers.json", zap.Error(err))
		return nil, err
	}

	zap.L().Info("New server added", zap.Int("server_id", server.Id))

	GetClient(server)
	handlers.GetListenerSocket(server.Id)

	ctx, cancel := context.WithCancel(context.Background())
	server.Ctx = ctx
	server.CancelFunc = cancel

	go StartReconnectLoop(ctx, server)

	handlers.BroadcastServers(config.AppEnv.Servers.ToServerResponses())

	return server, nil
}

func DeleteServer(serverId int) error {
	for i, server := range config.AppEnv.Servers {
		if server.Id == serverId {
			config.AppEnv.Servers = slices.Delete(config.AppEnv.Servers, i, i+1)
			if err := lib.WriteFile("./servers.json", &config.AppEnv.Servers); err != nil {
				zap.L().Error("Failed to write servers.json", zap.Error(err))
				return err
			}
			zap.L().Info("Server deleted", zap.Int("server_id", serverId))
			handlers.BroadcastServers(config.AppEnv.Servers.ToServerResponses())
			ShutdownServer(server)
			return nil
		}
	}
	return errors.New("server not found")
}

func UpdateServer(serverId int, server *structs.Server) (*structs.Server, error) {
	for i, s := range config.AppEnv.Servers {
		if s.Id == serverId {
			server.Id = serverId
			server.IsLocal = lib.IsLocalHostname(server.Host)
			config.AppEnv.Servers[i] = server
			if err := lib.WriteFile("./servers.json", &config.AppEnv.Servers); err != nil {
				zap.L().Error("Failed to write servers.json", zap.Error(err))
				return nil, err
			}
			zap.L().Info("Server updated", zap.Int("server_id", serverId))
			ShutdownServer(s)
			GetClient(server)
			handlers.GetListenerSocket(server.Id)
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
