package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kodra-pay/merchant-service/internal/config"
	"github.com/kodra-pay/merchant-service/internal/handlers"
	"github.com/kodra-pay/merchant-service/internal/repositories"
)

func Register(app *fiber.App, cfg config.Config, repo *repositories.MerchantRepository) {
	health := handlers.NewHealthHandler(cfg.ServiceName)
	health.Register(app)

	// TODO: add more routes here
}
