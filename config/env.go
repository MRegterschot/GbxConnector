package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/MRegterschot/GbxConnector/structs"
	"github.com/MRegterschot/GbxConnector/utils"
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
	if err = utils.ReadFile("./servers.json", &servers); err != nil {
		return err
	}

	// Check if the server is local
	for i, server := range servers {
		servers[i].IsLocal = utils.IsLocalHostname(server.Host)
	}

	AppEnv = &structs.Env{
		Port:       port,
		CorsOrigin: os.Getenv("CORS_ORIGIN"),
		Servers:    servers,
	}

	fmt.Println(AppEnv.Servers)

	return nil
}
