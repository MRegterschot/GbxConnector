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
		if err = lib.CreateIfNotExists("./servers.json"); err != nil {
			return err
		}
	}

	for _, server := range servers {
		server.Info = &structs.ServerInfo{
			LiveInfo: &structs.LiveInfo{
				Teams:   make(map[int]structs.Team),
				Players: make(map[string]structs.PlayerRound),
			},
		}
	}

	reconnectInterval, err := strconv.Atoi(os.Getenv("SERVER_RECONNECT_INTERVAL"))
	if err != nil {
		reconnectInterval = 5
	}

	AppEnv = &structs.Env{
		Port:              port,
		CorsOrigins:       strings.Split(os.Getenv("CORS_ORIGINS"), ","),
		LogLevel:          os.Getenv("LOG_LEVEL"),
		JwtSecret:         os.Getenv("JWT_SECRET"),
		InternalApiKey:    os.Getenv("INTERNAL_API_KEY"),
		ReconnectInterval: time.Duration(reconnectInterval) * time.Second,
		Servers:           servers,
	}

	return nil
}
