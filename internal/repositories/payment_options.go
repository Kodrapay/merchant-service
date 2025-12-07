package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/kodra-pay/merchant-service/internal/models"
	"github.com/lib/pq"
)

type PaymentOptionsRepository struct {
	db *sql.DB
}

func NewPaymentOptionsRepository(db *sql.DB) *PaymentOptionsRepository {
	return &PaymentOptionsRepository{db: db}
}

// GetByMerchantID retrieves payment options for a merchant
func (r *PaymentOptionsRepository) GetByMerchantID(ctx context.Context, merchantID int) (*models.PaymentOptions, error) {
	query := `
		SELECT id, merchant_id, card_enabled, bank_transfer_enabled, ussd_enabled,
		       qr_enabled, mobile_money_enabled, card_fee_type, card_fee_value, card_cap,
		       bank_transfer_fee_type, bank_transfer_fee_value, ussd_fee_type, ussd_fee_value,
		       created_at, updated_at
		FROM payment_options
		WHERE merchant_id = $1
	`

	var po models.PaymentOptions
	err := r.db.QueryRowContext(ctx, query, merchantID).Scan(
		&po.ID, &po.MerchantID, &po.CardEnabled, &po.BankTransferEnabled,
		&po.USSDEnabled, &po.QREnabled, &po.MobileMoneyEnabled,
		&po.CardFeeType, &po.CardFeeValue, &po.CardCap,
		&po.BankTransferFeeType, &po.BankTransferFeeValue,
		&po.USSDFeeType, &po.USSDFeeValue,
		&po.CreatedAt, &po.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		// Return default payment options if none exist
		return r.CreateDefault(ctx, merchantID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get payment options: %w", err)
	}

	return &po, nil
}

// CreateDefault creates default payment options for a merchant
func (r *PaymentOptionsRepository) CreateDefault(ctx context.Context, merchantID int) (*models.PaymentOptions, error) {
	query := `
		INSERT INTO payment_options (
			merchant_id, card_enabled, bank_transfer_enabled, ussd_enabled,
			card_fee_type, card_fee_value, bank_transfer_fee_type, bank_transfer_fee_value
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, merchant_id, card_enabled, bank_transfer_enabled, ussd_enabled,
		          qr_enabled, mobile_money_enabled, card_fee_type, card_fee_value, card_cap,
		          bank_transfer_fee_type, bank_transfer_fee_value, ussd_fee_type, ussd_fee_value,
		          created_at, updated_at
	`

	var po models.PaymentOptions
	err := r.db.QueryRowContext(
		ctx, query,
		merchantID, true, true, false,
		models.FeeTypePercentage, 1.5,
		models.FeeTypeFlat, 100,
	).Scan(
		&po.ID, &po.MerchantID, &po.CardEnabled, &po.BankTransferEnabled,
		&po.USSDEnabled, &po.QREnabled, &po.MobileMoneyEnabled,
		&po.CardFeeType, &po.CardFeeValue, &po.CardCap,
		&po.BankTransferFeeType, &po.BankTransferFeeValue,
		&po.USSDFeeType, &po.USSDFeeValue,
		&po.CreatedAt, &po.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create default payment options: %w", err)
	}

	return &po, nil
}

// Update updates payment options for a merchant
func (r *PaymentOptionsRepository) Update(ctx context.Context, po *models.PaymentOptions) error {
	query := `
		UPDATE payment_options SET
			card_enabled = $2,
			bank_transfer_enabled = $3,
			ussd_enabled = $4,
			qr_enabled = $5,
			mobile_money_enabled = $6,
			card_fee_type = $7,
			card_fee_value = $8,
			card_cap = $9,
			bank_transfer_fee_type = $10,
			bank_transfer_fee_value = $11,
			ussd_fee_type = $12,
			ussd_fee_value = $13,
			updated_at = NOW()
		WHERE merchant_id = $1
	`

	result, err := r.db.ExecContext(
		ctx, query,
		po.MerchantID, po.CardEnabled, po.BankTransferEnabled, po.USSDEnabled,
		po.QREnabled, po.MobileMoneyEnabled, po.CardFeeType, po.CardFeeValue,
		po.CardCap, po.BankTransferFeeType, po.BankTransferFeeValue,
		po.USSDFeeType, po.USSDFeeValue,
	)

	if err != nil {
		return fmt.Errorf("failed to update payment options: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("payment options not found for merchant: %d", po.MerchantID)
	}

	return nil
}

// SettlementConfigRepository handles settlement configuration operations
type SettlementConfigRepository struct {
	db *sql.DB
}

func NewSettlementConfigRepository(db *sql.DB) *SettlementConfigRepository {
	return &SettlementConfigRepository{db: db}
}

// GetByMerchantID retrieves settlement config for a merchant
func (r *SettlementConfigRepository) GetByMerchantID(ctx context.Context, merchantID int) (*models.SettlementConfig, error) {
	query := `
		SELECT id, merchant_id, schedule_type, settlement_time, settlement_days,
		       minimum_amount, auto_settle, settlement_delay_days, currency,
		       created_at, updated_at
		FROM settlement_configs
		WHERE merchant_id = $1
	`

	var sc models.SettlementConfig
	var settlementDays pq.Int64Array

	err := r.db.QueryRowContext(ctx, query, merchantID).Scan(
		&sc.ID, &sc.MerchantID, &sc.ScheduleType, &sc.SettlementTime,
		&settlementDays, &sc.MinimumAmount, &sc.AutoSettle,
		&sc.SettlementDelayDays, &sc.Currency, &sc.CreatedAt, &sc.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		// Return default settlement config if none exist
		return r.CreateDefault(ctx, merchantID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get settlement config: %w", err)
	}

	// Convert []int64 to []int
	sc.SettlementDays = make([]int, len(settlementDays))
	for i, v := range settlementDays {
		sc.SettlementDays[i] = int(v)
	}

	return &sc, nil
}

// CreateDefault creates default settlement config for a merchant
func (r *SettlementConfigRepository) CreateDefault(ctx context.Context, merchantID int) (*models.SettlementConfig, error) {
	query := `
		INSERT INTO settlement_configs (
			merchant_id, schedule_type, settlement_time, settlement_days,
			minimum_amount, auto_settle, settlement_delay_days, currency
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, merchant_id, schedule_type, settlement_time, settlement_days,
		          minimum_amount, auto_settle, settlement_delay_days, currency,
		          created_at, updated_at
	`

	defaultDays := pq.Int64Array{1, 2, 3, 4, 5} // Monday to Friday

	var sc models.SettlementConfig
	var settlementDays pq.Int64Array

	err := r.db.QueryRowContext(
		ctx, query,
		merchantID, models.ScheduleTypeDaily, "09:00:00", defaultDays,
		1000000, true, 2, "NGN",
	).Scan(
		&sc.ID, &sc.MerchantID, &sc.ScheduleType, &sc.SettlementTime,
		&settlementDays, &sc.MinimumAmount, &sc.AutoSettle,
		&sc.SettlementDelayDays, &sc.Currency, &sc.CreatedAt, &sc.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create default settlement config: %w", err)
	}

	// Convert []int64 to []int
	sc.SettlementDays = make([]int, len(settlementDays))
	for i, v := range settlementDays {
		sc.SettlementDays[i] = int(v)
	}

	return &sc, nil
}

// Update updates settlement config for a merchant
func (r *SettlementConfigRepository) Update(ctx context.Context, sc *models.SettlementConfig) error {
	// Convert []int to pq.Int64Array
	settlementDays := make(pq.Int64Array, len(sc.SettlementDays))
	for i, v := range sc.SettlementDays {
		settlementDays[i] = int64(v)
	}

	query := `
		UPDATE settlement_configs SET
			schedule_type = $2,
			settlement_time = $3,
			settlement_days = $4,
			minimum_amount = $5,
			auto_settle = $6,
			settlement_delay_days = $7,
			updated_at = NOW()
		WHERE merchant_id = $1
	`

	result, err := r.db.ExecContext(
		ctx, query,
		sc.MerchantID, sc.ScheduleType, sc.SettlementTime, settlementDays,
		sc.MinimumAmount, sc.AutoSettle, sc.SettlementDelayDays,
	)

	if err != nil {
		return fmt.Errorf("failed to update settlement config: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("settlement config not found for merchant: %d", sc.MerchantID)
	}

	return nil
}

// ListDueForSettlement retrieves all merchants that should be settled today
func (r *SettlementConfigRepository) ListDueForSettlement(ctx context.Context) ([]*models.SettlementConfig, error) {
	// Get current day of week (1=Mon, 7=Sun)
	dayOfWeek := int(ctx.Value("day_of_week").(int))
	if dayOfWeek == 0 {
		dayOfWeek = 7
	}

	query := `
		SELECT id, merchant_id, schedule_type, settlement_time, settlement_days,
		       minimum_amount, auto_settle, settlement_delay_days, currency,
		       created_at, updated_at
		FROM settlement_configs
		WHERE auto_settle = true
		  AND (
		      schedule_type = 'daily'
		      OR (schedule_type = 'weekly' AND $1 = ANY(settlement_days))
		  )
	`

	rows, err := r.db.QueryContext(ctx, query, dayOfWeek)
	if err != nil {
		return nil, fmt.Errorf("failed to list settlements due: %w", err)
	}
	defer rows.Close()

	var configs []*models.SettlementConfig
	for rows.Next() {
		var sc models.SettlementConfig
		var settlementDays pq.Int64Array

		err := rows.Scan(
			&sc.ID, &sc.MerchantID, &sc.ScheduleType, &sc.SettlementTime,
			&settlementDays, &sc.MinimumAmount, &sc.AutoSettle,
			&sc.SettlementDelayDays, &sc.Currency, &sc.CreatedAt, &sc.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan settlement config: %w", err)
		}

		// Convert []int64 to []int
		sc.SettlementDays = make([]int, len(settlementDays))
		for i, v := range settlementDays {
			sc.SettlementDays[i] = int(v)
		}

		configs = append(configs, &sc)
	}

	return configs, nil
}
