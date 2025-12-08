package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/kodra-pay/merchant-service/internal/models"
	"github.com/kodra-pay/merchant-service/internal/services"
)

type PaymentOptionsHandler struct {
	paymentSvc    *services.PaymentOptionsService
	settlementSvc *services.SettlementConfigService
}

func NewPaymentOptionsHandler(
	paymentSvc *services.PaymentOptionsService,
	settlementSvc *services.SettlementConfigService,
) *PaymentOptionsHandler {
	return &PaymentOptionsHandler{
		paymentSvc:    paymentSvc,
		settlementSvc: settlementSvc,
	}
}

// GetPaymentOptions retrieves payment options for a merchant
// GET /merchants/:id/payment-options
func (h *PaymentOptionsHandler) GetPaymentOptions(c *fiber.Ctx) error {
	merchantIDStr := c.Params("id")
	if merchantIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "merchant_id is required",
		})
	}
	merchantID, err := strconv.Atoi(merchantIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "merchant_id must be a number",
		})
	}

	options, err := h.paymentSvc.GetPaymentOptions(c.Context(), merchantID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(options)
}

// UpdatePaymentOptions updates payment options for a merchant
// PUT /merchants/:id/payment-options
func (h *PaymentOptionsHandler) UpdatePaymentOptions(c *fiber.Ctx) error {
	merchantIDStr := c.Params("id")
	if merchantIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "merchant_id is required",
		})
	}
	merchantID, err := strconv.Atoi(merchantIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "merchant_id must be a number",
		})
	}

	var req models.PaymentOptions
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	req.MerchantID = merchantID

	if err := h.paymentSvc.UpdatePaymentOptions(c.Context(), merchantID, &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "payment options updated successfully",
	})
}

// GetSettlementConfig retrieves settlement config for a merchant
// GET /merchants/:id/settlement-config
func (h *PaymentOptionsHandler) GetSettlementConfig(c *fiber.Ctx) error {
	merchantIDStr := c.Params("id")
	if merchantIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "merchant_id is required",
		})
	}
	merchantID, err := strconv.Atoi(merchantIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "merchant_id must be a number",
		})
	}

	config, err := h.settlementSvc.GetSettlementConfig(c.Context(), merchantID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(config)
}

// UpdateSettlementConfig updates settlement config for a merchant
// PUT /merchants/:id/settlement-config
func (h *PaymentOptionsHandler) UpdateSettlementConfig(c *fiber.Ctx) error {
	merchantIDStr := c.Params("id")
	if merchantIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "merchant_id is required",
		})
	}
	merchantID, err := strconv.Atoi(merchantIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "merchant_id must be a number",
		})
	}

	var req models.SettlementConfig
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	req.MerchantID = merchantID

	if err := h.settlementSvc.UpdateSettlementConfig(c.Context(), merchantID, &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "settlement config updated successfully",
	})
}

// Register registers all payment options routes
func (h *PaymentOptionsHandler) Register(app *fiber.App) {
	merchants := app.Group("/merchants")
	merchants.Get("/:id/payment-options", h.GetPaymentOptions)
	merchants.Put("/:id/payment-options", h.UpdatePaymentOptions)
	merchants.Get("/:id/settlement-config", h.GetSettlementConfig)
	merchants.Put("/:id/settlement-config", h.UpdateSettlementConfig)
}
