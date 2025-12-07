package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kodra-pay/merchant-service/internal/dto"
	"github.com/kodra-pay/merchant-service/internal/services"
)

type KYCHandler struct {
	merchantService *services.MerchantService
	kycService      *services.KYCService
}

func NewKYCHandler(merchantService *services.MerchantService, kycService *services.KYCService) *KYCHandler {
	return &KYCHandler{
		merchantService: merchantService,
		kycService:      kycService,
	}
}

// SubmitKYC handles KYC submission requests
func (h *KYCHandler) SubmitKYC(c *fiber.Ctx) error {
	var req dto.KYCSubmissionRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	submission, err := h.kycService.Submit(c.Context(), req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(submission)
}

// GetKYCStatus returns the current KYC status for a merchant
func (h *KYCHandler) GetKYCStatus(c *fiber.Ctx) error {
	merchantID, err := c.ParamsInt("merchant_id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid merchant ID")
	}

	status, err := h.kycService.GetLatest(c.Context(), merchantID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to fetch kyc status")
	}

	// If no KYC submission found, get the merchant's KYC status from merchants table
	if status == nil || status.MerchantID == 0 { // Check if status is nil or has a zero ID
		merchant, err := h.merchantService.GetMerchant(c.Context(), merchantID)
		if err != nil {
			return fiber.NewError(fiber.StatusNotFound, "merchant not found")
		}

		// Return merchant's KYC status from the merchants table
		return c.JSON(dto.KYCStatusResponse{
			MerchantID: merchantID,
			Status:     string(merchant.KYCStatus),
		})
	}

	return c.JSON(status)
}

// UpdateKYCStatus updates the KYC status (admin only)
func (h *KYCHandler) UpdateKYCStatus(c *fiber.Ctx) error {
	var req dto.KYCStatusUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if req.MerchantID == 0 { // int check
		return fiber.NewError(fiber.StatusBadRequest, "merchant_id is required")
	}

	// Convert req.ReviewerID (int) to *int for service layer
	var reviewerID *int
	if req.ReviewerID != 0 {
		reviewerID = &req.ReviewerID
	}

	if err := h.kycService.UpdateStatus(c.Context(), req.MerchantID, req.Status, reviewerID, &req.ReviewNotes); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return c.JSON(dto.KYCStatusResponse{
		MerchantID:  req.MerchantID,
		Status:      req.Status,
		ReviewerID:  reviewerID,
		ReviewNotes: req.ReviewNotes,
	})
}

// Register registers the KYC routes
func (h *KYCHandler) Register(app *fiber.App) {
	kyc := app.Group("/kyc")
	kyc.Post("/submit", h.SubmitKYC)
	kyc.Get("/status/:merchant_id", h.GetKYCStatus)
	kyc.Post("/update", h.UpdateKYCStatus) // Admin only - should be protected by auth middleware
	kyc.Get("/pending", h.ListPending)
}

func (h *KYCHandler) ListPending(c *fiber.Ctx) error {
	items, err := h.kycService.ListByStatus(c.Context(), "pending", 100)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to list kyc submissions")
	}
	return c.JSON(items)
}
