package config

import (
	"os"
	"strconv"
	"strings"
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
		return err
	}

	// Check if the server is local and check if we can connect
	for i, server := range servers {
		servers[i].IsLocal = lib.IsLocalHostname(server.Host)
	}

	reconnectInterval, err := strconv.Atoi(os.Getenv("SERVER_RECONNECT_INTERVAL"))
	if err != nil {
		reconnectInterval = 5
	}

	AppEnv = &structs.Env{
		Port:              port,
		CorsOrigins:       strings.Split(os.Getenv("CORS_ORIGINS"), ","),
		LogLevel:          os.Getenv("LOG_LEVEL"),
		ReconnectInterval: time.Duration(reconnectInterval) * time.Second,
		Servers:           servers,
	}

	return nil
}
