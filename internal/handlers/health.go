package handlers

import (
	"github.com/gofiber/fiber/v2"
)

type HealthHandler struct {
	serviceName string
}

func NewHealthHandler(serviceName string) *HealthHandler {
	return &HealthHandler{
		serviceName: serviceName,
	}
}

func (h *HealthHandler) Register(app *fiber.App) {
	app.Get("/health", h.Health)
}

func (h *HealthHandler) Health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "ok",
		"service": h.serviceName,
	})
}
