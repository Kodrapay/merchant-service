package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kodra-pay/merchant-service/internal/services"
)

type BalanceHandler struct {
	svc *services.BalanceService
}

func NewBalanceHandler(svc *services.BalanceService) *BalanceHandler {
	return &BalanceHandler{svc: svc}
}

// GetBalance returns the merchant's balance
func (h *BalanceHandler) GetBalance(c *fiber.Ctx) error {
	merchantID := c.Params("id")
	if merchantID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "merchant_id is required")
	}

	currency := c.Query("currency", "NGN")
	balance := h.svc.GetBalance(c.Context(), merchantID, currency)
	return c.JSON(balance)
}

// Register registers balance routes
func (h *BalanceHandler) Register(app *fiber.App) {
	app.Get("/merchants/:id/balance", h.GetBalance)
}
