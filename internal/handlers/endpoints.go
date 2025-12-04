package handlers

import (
	"github.com/gofiber/fiber/v2"
	"strings"

	"github.com/kodra-pay/merchant-service/internal/dto"
	"github.com/kodra-pay/merchant-service/internal/models"
	"github.com/kodra-pay/merchant-service/internal/services"
)

type MerchantHandler struct {
	svc *services.MerchantService
}

func NewMerchantHandler(svc *services.MerchantService) *MerchantHandler {
	return &MerchantHandler{svc: svc}
}

func (h *MerchantHandler) List(c *fiber.Ctx) error {
	resp := h.svc.List(c.Context())
	return c.JSON(resp)
}

func (h *MerchantHandler) ListMerchantsByKYCStatuses(c *fiber.Ctx) error {
	kycStatusesStr := c.Query("kyc_status")
	limit := c.QueryInt("limit", 100)
	offset := c.QueryInt("offset", 0)

	var kycStatuses []models.KYCStatus
	if kycStatusesStr != "" {
		for _, s := range splitCommaSeparatedString(kycStatusesStr) {
			kycStatuses = append(kycStatuses, models.KYCStatus(s))
		}
	}

	resp := h.svc.ListByKYCStatuses(c.Context(), kycStatuses, limit, offset)
	return c.JSON(resp)
}

func splitCommaSeparatedString(s string) []string {
	var result []string
	parts := strings.Split(s, ",")
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func (h *MerchantHandler) Create(c *fiber.Ctx) error {
	var req dto.MerchantCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	resp := h.svc.Create(c.Context(), req)
	return c.JSON(resp)
}

func (h *MerchantHandler) Get(c *fiber.Ctx) error {
	id := c.Params("id")
	resp := h.svc.Get(c.Context(), id)
	return c.JSON(resp)
}

func (h *MerchantHandler) UpdateStatus(c *fiber.Ctx) error {
	id := c.Params("id")
	var req dto.MerchantStatusUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	resp := h.svc.UpdateStatus(c.Context(), id, req)
	return c.JSON(resp)
}

func (h *MerchantHandler) UpdateKYCStatus(c *fiber.Ctx) error {
	id := c.Params("id")
	var req dto.MerchantKYCStatusUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	resp := h.svc.UpdateKYCStatus(c.Context(), id, req)
	return c.JSON(resp)
}

func (h *MerchantHandler) ListAPIKeys(c *fiber.Ctx) error {
	id := c.Params("id")
	resp := h.svc.ListAPIKeys(c.Context(), id)
	return c.JSON(resp)
}

func (h *MerchantHandler) RotateAPIKey(c *fiber.Ctx) error {
	id := c.Params("id")
	resp := h.svc.RotateAPIKey(c.Context(), id)
	return c.JSON(resp)
}

// Register registers all merchant routes
func (h *MerchantHandler) Register(app *fiber.App) {
	merchants := app.Group("/merchants")
	merchants.Get("/", h.List)
	merchants.Get("/kyc", h.ListMerchantsByKYCStatuses) // New route for listing by KYC status
	merchants.Post("/", h.Create)
	merchants.Get("/:id", h.Get)
	merchants.Put("/:id/status", h.UpdateStatus)
	merchants.Put("/:id/kyc-status", h.UpdateKYCStatus) // New route for updating KYC status
	merchants.Get("/:id/api-keys", h.ListAPIKeys)
	merchants.Post("/:id/api-keys/rotate", h.RotateAPIKey)
}
