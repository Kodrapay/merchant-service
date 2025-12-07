package services

import (
	"context"
	"fmt"

	"github.com/kodra-pay/merchant-service/internal/models"
	"github.com/kodra-pay/merchant-service/internal/repositories"
)

type PaymentOptionsService struct {
	repo *repositories.PaymentOptionsRepository
}

func NewPaymentOptionsService(repo *repositories.PaymentOptionsRepository) *PaymentOptionsService {
	return &PaymentOptionsService{repo: repo}
}

// GetPaymentOptions retrieves payment options for a merchant
func (s *PaymentOptionsService) GetPaymentOptions(ctx context.Context, merchantID int) (*models.PaymentOptions, error) {
	return s.repo.GetByMerchantID(ctx, merchantID)
}

// UpdatePaymentOptions updates payment options for a merchant
func (s *PaymentOptionsService) UpdatePaymentOptions(ctx context.Context, merchantID int, po *models.PaymentOptions) error {
	// Validate merchant owns this configuration
	if po.MerchantID != merchantID {
		return fmt.Errorf("merchant ID mismatch")
	}

	// Validate at least one payment method is enabled
	if !po.CardEnabled && !po.BankTransferEnabled && !po.USSDEnabled && !po.QREnabled && !po.MobileMoneyEnabled {
		return fmt.Errorf("at least one payment method must be enabled")
	}

	// Validate fee values
	if po.CardEnabled {
		if po.CardFeeType == models.FeeTypePercentage && (po.CardFeeValue < 0 || po.CardFeeValue > 100) {
			return fmt.Errorf("card fee percentage must be between 0 and 100")
		}
		if po.CardFeeType == models.FeeTypeFlat && po.CardFeeValue < 0 {
			return fmt.Errorf("card fee value must be positive")
		}
	}

	return s.repo.Update(ctx, po)
}

// SettlementConfigService handles settlement configuration operations
type SettlementConfigService struct {
	repo *repositories.SettlementConfigRepository
}

func NewSettlementConfigService(repo *repositories.SettlementConfigRepository) *SettlementConfigService {
	return &SettlementConfigService{repo: repo}
}

// GetSettlementConfig retrieves settlement config for a merchant
func (s *SettlementConfigService) GetSettlementConfig(ctx context.Context, merchantID int) (*models.SettlementConfig, error) {
	return s.repo.GetByMerchantID(ctx, merchantID)
}

// UpdateSettlementConfig updates settlement config for a merchant
func (s *SettlementConfigService) UpdateSettlementConfig(ctx context.Context, merchantID int, sc *models.SettlementConfig) error {
	// Validate merchant owns this configuration
	if sc.MerchantID != merchantID {
		return fmt.Errorf("merchant ID mismatch")
	}

	// Validate schedule type
	if sc.ScheduleType != models.ScheduleTypeDaily &&
		sc.ScheduleType != models.ScheduleTypeWeekly &&
		sc.ScheduleType != models.ScheduleTypeManual {
		return fmt.Errorf("invalid schedule type: %s", sc.ScheduleType)
	}

	// Validate settlement days for weekly schedules
	if sc.ScheduleType == models.ScheduleTypeWeekly {
		if len(sc.SettlementDays) == 0 {
			return fmt.Errorf("weekly schedule must have at least one settlement day")
		}
		for _, day := range sc.SettlementDays {
			if day < 1 || day > 7 {
				return fmt.Errorf("settlement day must be between 1 (Monday) and 7 (Sunday)")
			}
		}
	}

	// Validate minimum amount
	if sc.MinimumAmount < 0 {
		return fmt.Errorf("minimum amount must be positive")
	}

	// Validate settlement delay
	if sc.SettlementDelayDays < 0 || sc.SettlementDelayDays > 30 {
		return fmt.Errorf("settlement delay must be between 0 and 30 days")
	}

	return s.repo.Update(ctx, sc)
}

// ListDueForSettlement retrieves all merchants that should be settled today
func (s *SettlementConfigService) ListDueForSettlement(ctx context.Context) ([]*models.SettlementConfig, error) {
	return s.repo.ListDueForSettlement(ctx)
}
