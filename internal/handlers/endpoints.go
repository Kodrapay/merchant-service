package handlers

import (
	"github.com/gofiber/fiber/v2"

	"github.com/kodra-pay/merchant-service/internal/dto"
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
	merchants.Post("/", h.Create)
	merchants.Get("/:id", h.Get)
	merchants.Put("/:id/status", h.UpdateStatus)
	merchants.Get("/:id/api-keys", h.ListAPIKeys)
	merchants.Post("/:id/api-keys/rotate", h.RotateAPIKey)
}
