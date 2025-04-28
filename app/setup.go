package app

import (
	"strconv"

	"github.com/MRegterschot/GbxConnector/config"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"go.uber.org/zap"
)

func SetupAndRunApp() error {
	// load env
	err := config.LoadEnv()
	if err != nil {
		return err
	}

	config.SetupLogger()

	defer zap.L().Sync()

	// create app
	app := fiber.New(fiber.Config{
		BodyLimit: 4 * 1024 * 1024, // 4 MB
	})

	// attach middleware
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${ip}]:${port} ${status} - ${method} ${path} ${latency}\n",
	}))

	// setup routes
	SetupRoutes(app)

	app.Listen(":" + strconv.Itoa(config.AppEnv.Port))

	return nil
}
