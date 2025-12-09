package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/kodra-pay/merchant-service/internal/models"
)

var ErrPaymentLinkNotFound = errors.New("payment link not found")

type PaymentLinkRepository struct {
	db *sql.DB
}

func NewPaymentLinkRepository(db *sql.DB) *PaymentLinkRepository {
	return &PaymentLinkRepository{db: db}
}

func (r *PaymentLinkRepository) Create(ctx context.Context, link *models.PaymentLink) error {
	query := `
		INSERT INTO payment_links (merchant_id, mode, amount, currency, description, status, signature)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, reference, created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		link.MerchantID,
		link.Mode,
		link.Amount,
		link.Currency,
		link.Description,
		link.Status,
		link.Signature,
	).Scan(&link.ID, &link.Reference, &link.CreatedAt, &link.UpdatedAt)
}

func (r *PaymentLinkRepository) GetByID(ctx context.Context, id int) (*models.PaymentLink, error) {
	query := `
		SELECT id, merchant_id, mode, amount, currency, description, reference, status, signature, expires_at, created_at, updated_at
		FROM payment_links
		WHERE id = $1
	`
	var (
		link    models.PaymentLink
		amount  sql.NullInt64
		expires sql.NullTime
	)
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&link.ID,
		&link.MerchantID,
		&link.Mode,
		&amount,
		&link.Currency,
		&link.Description,
		&link.Reference,
		&link.Status,
		&link.Signature,
		&expires,
		&link.CreatedAt,
		&link.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if amount.Valid {
		val := amount.Int64
		link.Amount = &val
	}
	if expires.Valid {
		link.ExpiresAt = &expires.Time
	}
	return &link, nil
}

func (r *PaymentLinkRepository) GetByMerchantID(ctx context.Context, merchantID int, limit int) ([]models.PaymentLink, error) {
	query := `
		SELECT id, merchant_id, mode, amount, currency, description, reference, status, signature, expires_at, created_at, updated_at
		FROM payment_links
		WHERE merchant_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`
	rows, err := r.db.QueryContext(ctx, query, merchantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []models.PaymentLink
	for rows.Next() {
		var (
			link    models.PaymentLink
			amount  sql.NullInt64
			expires sql.NullTime
		)
		if err := rows.Scan(
			&link.ID,
			&link.MerchantID,
			&link.Mode,
			&amount,
			&link.Currency,
			&link.Description,
			&link.Reference,
			&link.Status,
			&link.Signature,
			&expires,
			&link.CreatedAt,
			&link.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if amount.Valid {
			val := amount.Int64
			link.Amount = &val
		}
		if expires.Valid {
			link.ExpiresAt = &expires.Time
		}
		links = append(links, link)
	}
	return links, rows.Err()
}

func (r *PaymentLinkRepository) Update(ctx context.Context, link *models.PaymentLink) error {
	query := `
		UPDATE payment_links
		SET status = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING updated_at
	`
	return r.db.QueryRowContext(ctx, query, link.Status, link.ID).Scan(&link.UpdatedAt)
}

// Delete removes a payment link by ID, optionally scoping to merchant ownership.
func (r *PaymentLinkRepository) Delete(ctx context.Context, id, merchantID int) error {
	var (
		res sql.Result
		err error
	)

	if merchantID != 0 {
		res, err = r.db.ExecContext(ctx, `
			DELETE FROM payment_links
			WHERE id = $1 AND merchant_id = $2
		`, id, merchantID)
	} else {
		res, err = r.db.ExecContext(ctx, `
			DELETE FROM payment_links
			WHERE id = $1
		`, id)
	}

	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrPaymentLinkNotFound
	}
	return nil
}
