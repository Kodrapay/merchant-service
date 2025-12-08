package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/kodra-pay/merchant-service/internal/dto"
	"github.com/kodra-pay/merchant-service/internal/repositories"
	"github.com/kodra-pay/merchant-service/internal/services"
)

type PaymentLinkHandler struct {
	svc *services.PaymentLinkService
}

func NewPaymentLinkHandler(svc *services.PaymentLinkService) *PaymentLinkHandler {
	return &PaymentLinkHandler{svc: svc}
}

func (h *PaymentLinkHandler) Register(app *fiber.App) {
	app.Post("/payment-links", h.CreatePaymentLink)
	app.Get("/payment-links/:id", h.GetPaymentLink)
	app.Get("/merchants/:merchant_id/payment-links", h.ListPaymentLinks)
	app.Delete("/payment-links/:id", h.DeletePaymentLink)
}

func (h *PaymentLinkHandler) CreatePaymentLink(c *fiber.Ctx) error {
	var req dto.CreatePaymentLinkRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	// Validate required fields
	if req.MerchantID <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "merchant_id is required")
	}
	if req.Currency == "" {
		return fiber.NewError(fiber.StatusBadRequest, "currency is required")
	}
	if req.Mode == "" {
		req.Mode = "fixed"
	}
	if req.Mode != "fixed" && req.Mode != "open" {
		return fiber.NewError(fiber.StatusBadRequest, "mode must be 'fixed' or 'open'")
	}
	if req.Mode == "fixed" && (req.Amount == nil || *req.Amount <= 0) {
		return fiber.NewError(fiber.StatusBadRequest, "amount is required for fixed mode")
	}

	link, err := h.svc.CreatePaymentLink(c.Context(), req)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to create payment link")
	}

	return c.Status(fiber.StatusCreated).JSON(link)
}

func (h *PaymentLinkHandler) GetPaymentLink(c *fiber.Ctx) error {
	idStr := c.Params("id")
	if idStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "payment link id is required")
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "payment link id must be a number")
	}

	link, err := h.svc.GetPaymentLink(c.Context(), id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to get payment link")
	}
	if link == nil {
		return fiber.NewError(fiber.StatusNotFound, "payment link not found")
	}

	return c.JSON(link)
}

func (h *PaymentLinkHandler) ListPaymentLinks(c *fiber.Ctx) error {
	merchantIDStr := c.Params("merchant_id")
	if merchantIDStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "merchant_id is required")
	}
	merchantID, err := strconv.Atoi(merchantIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "merchant_id must be a number")
	}

	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
			if limit > 100 {
				limit = 100
			}
		}
	}

	links, err := h.svc.ListPaymentLinks(c.Context(), merchantID, limit)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to list payment links")
	}

	return c.JSON(links)
}

func (h *PaymentLinkHandler) DeletePaymentLink(c *fiber.Ctx) error {
	idStr := c.Params("id")
	if idStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "payment link id is required")
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "payment link id must be a number")
	}
	merchantIDStr := c.Query("merchant_id")
	merchantID := 0
	if merchantIDStr != "" {
		merchantID, err = strconv.Atoi(merchantIDStr)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "merchant_id must be a number")
		}
	}

	if err := h.svc.DeletePaymentLink(c.Context(), id, merchantID); err != nil {
		if err == repositories.ErrPaymentLinkNotFound {
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Payment link deleted successfully",
	})
}
