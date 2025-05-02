package app

import (
	"context"
	"errors"

	"github.com/MRegterschot/GbxConnector/config"
	"github.com/MRegterschot/GbxConnector/handlers"
	"github.com/MRegterschot/GbxConnector/lib"
	"github.com/MRegterschot/GbxConnector/structs"
	"go.uber.org/zap"
)

func AddServer(server *structs.Server) error {
	if server == nil {
		return errors.New("server is nil")
	}

	server.IsLocal = lib.IsLocalHostname(server.Host)

	config.AppEnv.Servers = append(config.AppEnv.Servers, server)
	if err := lib.WriteFile("./servers.json", &config.AppEnv.Servers); err != nil {
		return err
	}

	GetClient(server)
	handlers.GetListenerSocket(server.Id)

	server.Ctx, server.CancelFunc = context.WithCancel(context.Background())

	go StartReconnectLoop(server.Ctx, server)

	zap.L().Info("Added server", zap.String("host", server.Host), zap.Int("port", server.XMLRPCPort))
	return nil
}
