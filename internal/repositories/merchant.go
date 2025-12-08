package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"log" // Added
	"time"

	_ "github.com/lib/pq"

	"github.com/kodra-pay/merchant-service/internal/models"
)

type MerchantRepository struct {
	db *sql.DB
}

func NewMerchantRepository(dsn string) (*MerchantRepository, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &MerchantRepository{db: db}, nil
}

func (r *MerchantRepository) Close() error {
	return r.db.Close()
}

func (r *MerchantRepository) GetDB() *sql.DB {
	return r.db
}

// Create inserts a new merchant
func (r *MerchantRepository) Create(ctx context.Context, merchant *models.Merchant) error {
	query := `
		INSERT INTO merchants (name, email, business_name, country, status, kyc_status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`
	var id int
	err := r.db.QueryRowContext(ctx, query,
		merchant.Name,
		merchant.Email,
		merchant.BusinessName,
		merchant.Country,
		merchant.Status,
		merchant.KYCStatus,
		merchant.CreatedAt,
		merchant.UpdatedAt,
	).Scan(&id) // Retrieve the generated ID

	if err == nil {
		merchant.ID = id // Update the merchant object with the ID from DB
	}
	return err
}

// GetByID retrieves a merchant by ID
func (r *MerchantRepository) GetByID(ctx context.Context, id int) (*models.Merchant, error) {
	query := `
		SELECT id, name, email, business_name, country, status, kyc_status, created_at, updated_at
		FROM merchants
		WHERE id = $1
	`

	merchant := &models.Merchant{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&merchant.ID,
		&merchant.Name,
		&merchant.Email,
		&merchant.BusinessName,
		&merchant.Country,
		&merchant.Status,
		&merchant.KYCStatus,
		&merchant.CreatedAt,
		&merchant.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("merchant not found")
	}

	return merchant, err
}

// GetByEmail retrieves a merchant by email
func (r *MerchantRepository) GetByEmail(ctx context.Context, email string) (*models.Merchant, error) {
	query := `
		SELECT id, name, email, business_name, country, status, kyc_status, created_at, updated_at
		FROM merchants
		WHERE email = $1
	`

	merchant := &models.Merchant{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&merchant.ID,
		&merchant.Name,
		&merchant.Email,
		&merchant.BusinessName,
		&merchant.Country,
		&merchant.Status,
		&merchant.KYCStatus,
		&merchant.CreatedAt,
		&merchant.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("merchant not found")
	}

	return merchant, err
}

// List retrieves all merchants with optional filters
func (r *MerchantRepository) List(ctx context.Context, status string, limit, offset int) ([]*models.Merchant, error) {
	query := `
		SELECT id, name, email, business_name, country, status, kyc_status, created_at, updated_at
		FROM merchants
	`

	args := []interface{}{}
	argCount := 1

	if status != "" {
		query += fmt.Sprintf(" WHERE status = $%d", argCount)
		args = append(args, status)
		argCount++
	}

	query += " ORDER BY created_at DESC"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, limit)
		argCount++
	}

	if offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	merchants := []*models.Merchant{}
	for rows.Next() {
		merchant := &models.Merchant{}
		err := rows.Scan(
			&merchant.ID,
			&merchant.Name,
			&merchant.Email,
			&merchant.BusinessName,
			&merchant.Country,
			&merchant.Status,
			&merchant.KYCStatus,
			&merchant.CreatedAt,
			&merchant.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		merchants = append(merchants, merchant)
	}

	return merchants, rows.Err()
}

// Update updates a merchant's information
func (r *MerchantRepository) Update(ctx context.Context, merchant *models.Merchant) error {
	query := `
		UPDATE merchants
		SET name = $2, business_name = $3, status = $4, kyc_status = $5, updated_at = $6
		WHERE id = $1
	`

	merchant.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		merchant.ID, // now int
		merchant.Name,
		merchant.BusinessName,
		merchant.Status,
		merchant.KYCStatus,
		merchant.UpdatedAt,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("merchant not found")
	}

	return nil
}

// UpdateStatus updates a merchant's status
func (r *MerchantRepository) UpdateStatus(ctx context.Context, id int, status models.MerchantStatus) error { // id changed to int
	query := `
		UPDATE merchants
		SET status = $2, updated_at = $3
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id, status, time.Now())
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("merchant not found")
	}

	return nil
}

// UpdateKYCStatus updates a merchant's KYC status
func (r *MerchantRepository) UpdateKYCStatus(ctx context.Context, id int, kycStatus models.KYCStatus) error { // id changed to int
	query := `
		UPDATE merchants
		SET kyc_status = $2, updated_at = $3
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id, kycStatus, time.Now())
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("merchant not found")
	}

	return nil
}

// Delete deletes a merchant
func (r *MerchantRepository) Delete(ctx context.Context, id int) error { // id changed to int
	query := `DELETE FROM merchants WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("merchant not found")
	}

	return nil
}

// Count returns the total number of merchants
func (r *MerchantRepository) Count(ctx context.Context, status string) (int, error) {
	query := `SELECT COUNT(*) FROM merchants`

	args := []interface{}{}
	if status != "" {
		query += " WHERE status = $1"
		args = append(args, status)
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

// ListByKYCStatus retrieves merchants by their KYC status
func (r *MerchantRepository) ListByKYCStatus(ctx context.Context, kycStatus models.KYCStatus, limit, offset int) ([]*models.Merchant, error) {
	query := `
		SELECT id, name, email, business_name, country, status, kyc_status, created_at, updated_at
		FROM merchants
		WHERE kyc_status = $1
	`

	args := []interface{}{kycStatus}
	argCount := 2

	query += " ORDER BY created_at DESC"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, limit)
		argCount++
	}

	if offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	merchants := []*models.Merchant{}
	for rows.Next() {
		merchant := &models.Merchant{}
		err := rows.Scan(
			&merchant.ID, // now int
			&merchant.Name,
			&merchant.Email,
			&merchant.BusinessName,
			&merchant.Country,
			&merchant.Status,
			&merchant.KYCStatus,
			&merchant.CreatedAt,
			&merchant.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		merchants = append(merchants, merchant)
	}

	return merchants, rows.Err()
}

// ListByKYCStatuses retrieves merchants by multiple KYC statuses
func (r *MerchantRepository) ListByKYCStatuses(ctx context.Context, kycStatuses []models.KYCStatus, limit, offset int) ([]*models.Merchant, error) {
	if len(kycStatuses) == 0 {
		return []*models.Merchant{}, nil
	}

	query := `
		SELECT id, name, email, business_name, country, status, kyc_status, created_at, updated_at
		FROM merchants
		WHERE kyc_status IN (`
	
	args := make([]interface{}, len(kycStatuses))
	for i, status := range kycStatuses {
		query += fmt.Sprintf("$%d", i+1)
		args[i] = status
		if i < len(kycStatuses)-1 {
			query += ", "
		}
	}
	query += ")"

	argCount := len(args) + 1

	query += " ORDER BY created_at DESC"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, limit)
		argCount++
	}

	if offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, offset)
	}

	log.Printf("DEBUG: ListByKYCStatuses query: %s, args: %v", query, args) // Added debug log

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	merchants := []*models.Merchant{}
	for rows.Next() {
		merchant := &models.Merchant{}
		err := rows.Scan(
			&merchant.ID, // now int
			&merchant.Name,
			&merchant.Email,
			&merchant.BusinessName,
			&merchant.Country,
			&merchant.Status,
			&merchant.KYCStatus,
			&merchant.CreatedAt,
			&merchant.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		merchants = append(merchants, merchant)
	}

	return merchants, rows.Err()
}

