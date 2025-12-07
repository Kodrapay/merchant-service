package repositories

import (
	"context"
	"database/sql"

	"github.com/kodra-pay/merchant-service/internal/models"
)

type APIKeyRepository struct {
	db *sql.DB
}

func NewAPIKeyRepository(db *sql.DB) *APIKeyRepository {
	return &APIKeyRepository{db: db}
}

func (r *APIKeyRepository) Create(ctx context.Context, key *models.APIKey) error {
	query := `
		INSERT INTO api_keys (merchant_id, key_hash, key_prefix, key_type, environment, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	return r.db.QueryRowContext(
		ctx,
		query,
		key.MerchantID,
		key.KeyHash,
		key.KeyPrefix,
		key.KeyType,
		key.Environment,
		key.IsActive,
		key.CreatedAt,
	).Scan(&key.ID)
}

func (r *APIKeyRepository) ListByMerchantID(ctx context.Context, merchantID int) ([]*models.APIKey, error) {
	query := `
		SELECT id, merchant_id, key_hash, key_prefix, key_type, environment, is_active, last_used_at, expires_at, created_at
		FROM api_keys
		WHERE merchant_id = $1 AND is_active = true
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, merchantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*models.APIKey
	for rows.Next() {
		key := &models.APIKey{}
		if err := rows.Scan(
			&key.ID,
			&key.MerchantID,
			&key.KeyHash,
			&key.KeyPrefix,
			&key.KeyType,
			&key.Environment,
			&key.IsActive,
			&key.LastUsedAt,
			&key.ExpiresAt,
			&key.CreatedAt,
		); err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}

	return keys, rows.Err()
}

func (r *APIKeyRepository) DeactivateByMerchantAndType(ctx context.Context, merchantID int, keyType models.APIKeyType, env models.Environment) error {
	query := `
		UPDATE api_keys
		SET is_active = false
		WHERE merchant_id = $1 AND key_type = $2 AND environment = $3
	`
	_, err := r.db.ExecContext(ctx, query, merchantID, keyType, env)
	return err
}

func (r *APIKeyRepository) GetByPrefix(ctx context.Context, keyPrefix string) (*models.APIKey, error) {
	query := `
		SELECT id, merchant_id, key_hash, key_prefix, key_type, environment, is_active, last_used_at, expires_at, created_at
		FROM api_keys
		WHERE key_prefix = $1 AND is_active = true
	`
	key := &models.APIKey{}
	err := r.db.QueryRowContext(ctx, query, keyPrefix).Scan(
		&key.ID,
		&key.MerchantID,
		&key.KeyHash,
		&key.KeyPrefix,
		&key.KeyType,
		&key.Environment,
		&key.IsActive,
		&key.LastUsedAt,
		&key.ExpiresAt,
		&key.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return key, nil
}
