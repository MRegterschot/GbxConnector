package config

import (
	"os"
	"strconv"
	"time"

	"github.com/MRegterschot/GbxConnector/lib"
	"github.com/MRegterschot/GbxConnector/structs"
	"github.com/joho/godotenv"
)

var AppEnv *structs.Env

func LoadEnv() error {
	_ = godotenv.Load()

	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		port = 6980
	}

	servers := make([]*structs.Server, 0)
	if err = lib.ReadFile("./servers.json", &servers); err != nil {
		if err = lib.CreateIfNotExists("./servers.json"); err != nil {
			return err
		}
	}

	for _, server := range servers {
		server.ResetLiveInfo()
	}

	reconnectInterval, err := strconv.Atoi(os.Getenv("SERVER_RECONNECT_INTERVAL"))
	if err != nil {
		reconnectInterval = 5
	}

	AppEnv = &structs.Env{
		Port:               port,
		LogLevel:           os.Getenv("LOG_LEVEL"),
		JwtSecret:          os.Getenv("JWT_SECRET"),
		ReconnectInterval:  time.Duration(reconnectInterval) * time.Second,
		DockerNetworkRange: os.Getenv("DOCKER_NETWORK_RANGE"),
		Servers:            servers,
	}

	return nil
}
