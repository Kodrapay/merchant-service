package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/kodra-pay/merchant-service/internal/models"
)

type BalanceRepository struct {
	db *sql.DB
}

func NewBalanceRepository(db *sql.DB) *BalanceRepository {
	return &BalanceRepository{db: db}
}

// GetOrCreate returns the merchant balance for a currency, creating it if it doesn't exist
func (r *BalanceRepository) GetOrCreate(ctx context.Context, merchantID int, currency string) (*models.MerchantBalance, error) {
	// Try to get existing balance
	var balance models.MerchantBalance
	err := r.db.QueryRowContext(ctx, `
		SELECT id, merchant_id, currency, pending_balance, available_balance, total_volume, created_at, updated_at
		FROM merchant_balances
		WHERE merchant_id = $1 AND currency = $2
	`, merchantID, currency).Scan(
		&balance.ID,
		&balance.MerchantID,
		&balance.Currency,
		&balance.PendingBalance,
		&balance.AvailableBalance,
		&balance.TotalVolume,
		&balance.CreatedAt,
		&balance.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		// Create new balance record
		err = r.db.QueryRowContext(ctx, `
			INSERT INTO merchant_balances (merchant_id, currency, pending_balance, available_balance, total_volume)
			VALUES ($1, $2, 0, 0, 0)
			RETURNING id, merchant_id, currency, pending_balance, available_balance, total_volume, created_at, updated_at
		`, merchantID, currency).Scan(
			&balance.ID,
			&balance.MerchantID,
			&balance.Currency,
			&balance.PendingBalance,
			&balance.AvailableBalance,
			&balance.TotalVolume,
			&balance.CreatedAt,
			&balance.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		return &balance, nil
	}

	if err != nil {
		return nil, err
	}

	return &balance, nil
}

// AddToPending adds amount to pending balance (when transaction succeeds)
func (r *BalanceRepository) AddToPending(ctx context.Context, merchantID int, currency string, amount int64) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO merchant_balances (merchant_id, currency, pending_balance, available_balance, total_volume)
		VALUES ($1, $2, $3, 0, $3)
		ON CONFLICT (merchant_id, currency)
		DO UPDATE SET
			pending_balance = merchant_balances.pending_balance + $3,
			total_volume = merchant_balances.total_volume + $3,
			updated_at = NOW()
	`, merchantID, currency, amount)
	return err
}

// SettlePending moves amount from pending to available (when settlement completes)
func (r *BalanceRepository) SettlePending(ctx context.Context, merchantID int, currency string, amount int64) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE merchant_balances
		SET pending_balance = pending_balance - $3,
			available_balance = available_balance + $3,
			updated_at = NOW()
		WHERE merchant_id = $1 AND currency = $2
		  AND pending_balance >= $3
	`, merchantID, currency, amount)
	if err != nil {
		return err
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("insufficient pending balance to settle %d", amount)
	}
	return nil
}

// DeductFromAvailable deducts amount from available balance (when payout is made)
func (r *BalanceRepository) DeductFromAvailable(ctx context.Context, merchantID int, currency string, amount int64) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE merchant_balances
		SET available_balance = available_balance - $3,
			updated_at = NOW()
		WHERE merchant_id = $1 AND currency = $2 AND available_balance >= $3
	`, merchantID, currency, amount)
	if err != nil {
		return err
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("insufficient available balance")
	}
	return nil
}
