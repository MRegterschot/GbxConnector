package app

import (
	"github.com/MRegterschot/GbxConnector/config"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func SetupRoutes(app *fiber.App) {
	// Middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins:     config.AppEnv.CorsOrigin,
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept",
		AllowCredentials: true,
	}))

	// Health check route
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})
}
