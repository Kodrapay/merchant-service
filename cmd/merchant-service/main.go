package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/kodra-pay/merchant-service/internal/config"
	"github.com/kodra-pay/merchant-service/internal/middleware"
	"github.com/kodra-pay/merchant-service/internal/routes"
)

func main() {
	cfg := config.Load("merchant-service", "7002")

	app := fiber.New()

	// Enable CORS for frontend access
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	app.Use(middleware.RequestID())

	routes.Register(app, cfg.ServiceName)

	log.Printf("%s listening on :%s", cfg.ServiceName, cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
