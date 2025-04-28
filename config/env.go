package config

import (
	"os"
	"strconv"

	"github.com/MRegterschot/GbxConnector/lib"
	"github.com/MRegterschot/GbxConnector/structs"
	"github.com/joho/godotenv"
)

var AppEnv *structs.Env

func LoadEnv() error {
	if err := godotenv.Load(); err != nil {
		return err
	}

	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		port = 6980
	}

	servers := make([]structs.Server, 0)
	if err = lib.ReadFile("./servers.json", &servers); err != nil {
		return err
	}

	// Check if the server is local
	for i, server := range servers {
		servers[i].IsLocal = lib.IsLocalHostname(server.Host)
	}

	AppEnv = &structs.Env{
		Port:       port,
		CorsOrigin: os.Getenv("CORS_ORIGIN"),
		LogLevel:   os.Getenv("LOG_LEVEL"),
		Servers:    servers,
	}

	return nil
}
