package handlers

import (
	"strconv"

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
	merchantIDStr := c.Params("id")
	if merchantIDStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "merchant_id is required")
	}
	merchantID, err := strconv.Atoi(merchantIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "merchant_id must be a number")
	}

	currency := c.Query("currency", "NGN")
	balance := h.svc.GetBalance(c.Context(), merchantID, currency)
	return c.JSON(balance)
}

// Register registers balance routes
func (h *BalanceHandler) Register(app *fiber.App) {
	app.Get("/merchants/:id/balance", h.GetBalance)
	app.Post("/internal/balance/settle", h.SettleBalance)
	app.Post("/internal/balance/record", h.RecordBalance)
}

// SettleBalance moves pending into available for a merchant (internal use)
func (h *BalanceHandler) SettleBalance(c *fiber.Ctx) error {
	var payload struct {
		MerchantID int    `json:"merchant_id"`
		Currency   string `json:"currency"`
		Amount     int64  `json:"amount"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if payload.MerchantID <= 0 || payload.Amount <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "merchant_id and positive amount are required")
	}
	if payload.Currency == "" {
		payload.Currency = "NGN"
	}

	if err := h.svc.Settle(c.Context(), payload.MerchantID, payload.Currency, payload.Amount); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// RecordBalance adds a successful transaction amount into pending balance
func (h *BalanceHandler) RecordBalance(c *fiber.Ctx) error {
	var payload struct {
		MerchantID int    `json:"merchant_id"`
		Currency   string `json:"currency"`
		Amount     int64  `json:"amount"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if payload.MerchantID <= 0 || payload.Amount <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "merchant_id and positive amount are required")
	}
	if payload.Currency == "" {
		payload.Currency = "NGN"
	}

	if err := h.svc.RecordTransaction(c.Context(), payload.MerchantID, payload.Currency, payload.Amount); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}
