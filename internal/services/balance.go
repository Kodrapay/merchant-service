package services

import (
	"context"

	"github.com/kodra-pay/merchant-service/internal/dto"
	"github.com/kodra-pay/merchant-service/internal/repositories"
)

type BalanceService struct {
	repo *repositories.BalanceRepository
}

func NewBalanceService(repo *repositories.BalanceRepository) *BalanceService {
	return &BalanceService{repo: repo}
}

// GetBalance returns the merchant's balance for a specific currency
func (s *BalanceService) GetBalance(ctx context.Context, merchantID int, currency string) dto.MerchantBalanceResponse {
	balance, err := s.repo.GetOrCreate(ctx, merchantID, currency)
	if err != nil {
		return dto.MerchantBalanceResponse{
			MerchantID:       merchantID,
			Currency:         currency,
			PendingBalance:   0,
			AvailableBalance: 0,
			TotalVolume:      0,
		}
	}

	return dto.MerchantBalanceResponse{
		MerchantID:       balance.MerchantID,
		Currency:         balance.Currency,
		PendingBalance:   balance.PendingBalance,
		AvailableBalance: balance.AvailableBalance,
		TotalVolume:      balance.TotalVolume,
	}
}

// RecordTransaction adds transaction amount to pending balance
func (s *BalanceService) RecordTransaction(ctx context.Context, merchantID int, currency string, amount int64) error {
	return s.repo.AddToPending(ctx, merchantID, currency, amount)
}

// SettleFunds moves funds from pending to available
func (s *BalanceService) SettleFunds(ctx context.Context, merchantID int, currency string, amount int64) error {
	return s.repo.SettlePending(ctx, merchantID, currency, amount)
}

// ProcessPayout deducts amount from available balance
func (s *BalanceService) ProcessPayout(ctx context.Context, merchantID int, currency string, amount int64) error {
	return s.repo.DeductFromAvailable(ctx, merchantID, currency, amount)
}

// Settle moves pending funds into available (used by settlement/payout flows)
func (s *BalanceService) Settle(ctx context.Context, merchantID int, currency string, amount int64) error {
	return s.repo.SettlePending(ctx, merchantID, currency, amount)
}
