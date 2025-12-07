package middleware

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/kodra-pay/merchant-service/internal/models"
	"github.com/kodra-pay/merchant-service/internal/services"
)

// KYCCheckMiddleware ensures that merchants have approved KYC before performing certain operations
type KYCCheckMiddleware struct {
	merchantService *services.MerchantService
}

func NewKYCCheckMiddleware(merchantService *services.MerchantService) *KYCCheckMiddleware {
	return &KYCCheckMiddleware{
		merchantService: merchantService,
	}
}

// RequireApprovedKYC is a middleware that checks if the merchant has approved KYC
func (m *KYCCheckMiddleware) RequireApprovedKYC(c *fiber.Ctx) error {
	// Extract merchant ID from context (assuming it's set by auth middleware)
	merchantID := c.Locals("merchant_id")
	if merchantID == nil {
		return fiber.NewError(fiber.StatusUnauthorized, "merchant not authenticated")
	}

	merchantIDStr, ok := merchantID.(string)
	if !ok {
		return fiber.NewError(fiber.StatusInternalServerError, "invalid merchant ID format")
	}
	id, err := strconv.Atoi(merchantIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "invalid merchant ID format")
	}

	// Get merchant details
	merchant, err := m.merchantService.GetMerchant(c.Context(), id)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "merchant not found")
	}

	// Check if merchant can transact
	if merchant.KYCStatus != models.KYCStatusApproved {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error":      "kyc_not_approved",
			"message":    "Your KYC verification must be approved before you can process transactions. Please complete your KYC verification.",
			"kyc_status": merchant.KYCStatus,
		})
	}

	if merchant.Status != models.MerchantStatusActive {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error":   "account_not_active",
			"message": "Your merchant account is not active. Please contact support.",
			"status":  merchant.Status,
		})
	}

	// Merchant is verified and active, allow request to proceed
	return c.Next()
}

// CheckKYCStatus returns the current KYC status without blocking the request
func (m *KYCCheckMiddleware) CheckKYCStatus(c *fiber.Ctx) error {
	merchantID := c.Locals("merchant_id")
	if merchantID == nil {
		return c.Next()
	}

	merchantIDStr, ok := merchantID.(string)
	if !ok {
		return c.Next()
	}
	id, err := strconv.Atoi(merchantIDStr)
	if err != nil {
		return c.Next()
	}

	merchant, err := m.merchantService.GetMerchant(c.Context(), id)
	if err != nil {
		return c.Next()
	}

	// Add KYC status to context for other handlers to use
	c.Locals("kyc_status", merchant.KYCStatus)
	c.Locals("can_transact", merchant.CanTransact)

	return c.Next()
}
