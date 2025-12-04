package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kodra-pay/merchant-service/internal/dto"
	"github.com/kodra-pay/merchant-service/internal/services"
)

type KYCHandler struct {
	merchantService *services.MerchantService
}

func NewKYCHandler(merchantService *services.MerchantService) *KYCHandler {
	return &KYCHandler{
		merchantService: merchantService,
	}
}

// SubmitKYC handles KYC submission requests
func (h *KYCHandler) SubmitKYC(c *fiber.Ctx) error {
	var req dto.KYCSubmissionRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	// TODO: Implement actual KYC submission logic
	// For now, return a stub response
	return c.JSON(dto.KYCSubmissionResponse{
		SubmissionID: "kyc_" + req.MerchantID,
		Status:       "pending",
		Message:      "KYC submission received and is under review",
	})
}

// GetKYCStatus returns the current KYC status for a merchant
func (h *KYCHandler) GetKYCStatus(c *fiber.Ctx) error {
	merchantID := c.Params("merchant_id")
	if merchantID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "merchant_id is required")
	}

	merchant, err := h.merchantService.GetMerchant(c.Context(), merchantID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "merchant not found")
	}

	return c.JSON(dto.KYCStatusResponse{
		MerchantID: merchant.ID,
		Status:     string(merchant.KYCStatus),
	})
}

// UpdateKYCStatus updates the KYC status (admin only)
func (h *KYCHandler) UpdateKYCStatus(c *fiber.Ctx) error {
	var req dto.KYCStatusUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	// TODO: Implement actual KYC status update logic
	// For now, return a stub response
	return c.JSON(dto.KYCStatusResponse{
		MerchantID:  req.MerchantID,
		Status:      req.Status,
		ReviewerID:  req.ReviewerID,
		ReviewNotes: req.ReviewNotes,
	})
}

// Register registers the KYC routes
func (h *KYCHandler) Register(app *fiber.App) {
	kyc := app.Group("/kyc")
	kyc.Post("/submit", h.SubmitKYC)
	kyc.Get("/status/:merchant_id", h.GetKYCStatus)
	kyc.Post("/update", h.UpdateKYCStatus) // Admin only - should be protected by auth middleware
}
