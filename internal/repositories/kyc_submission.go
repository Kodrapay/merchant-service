package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/kodra-pay/merchant-service/internal/models"
)

type KYCSubmissionRepository struct {
	db *sql.DB
}

func NewKYCSubmissionRepository(db *sql.DB) *KYCSubmissionRepository {
	return &KYCSubmissionRepository{db: db}
}

func (r *KYCSubmissionRepository) Create(ctx context.Context, submission *models.KYCSubmission) error {
	now := time.Now()
	submission.ID = uuid.NewString()
	submission.Status = "pending"
	submission.CreatedAt = now
	submission.UpdatedAt = now

	var documents interface{}
	if submission.Documents != nil {
		if b, err := json.Marshal(submission.Documents); err == nil {
			documents = b
		}
	}

	query := `
		INSERT INTO kyc_submissions (
			id, merchant_id, business_type, business_name, cac_number, tin_number,
			business_address, city, state, postal_code, incorporation_date,
			business_category, director_name, director_bvn, director_phone, director_email,
			documents, status, created_at, updated_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20
		)
	`

	_, err := r.db.ExecContext(ctx, query,
		submission.ID,
		submission.MerchantID,
		submission.BusinessType,
		submission.BusinessName,
		submission.CACNumber,
		submission.TINNumber,
		submission.BusinessAddress,
		submission.City,
		submission.State,
		submission.PostalCode,
		submission.IncorporationDate,
		submission.BusinessCategory,
		submission.DirectorName,
		submission.DirectorBVN,
		submission.DirectorPhone,
		submission.DirectorEmail,
		documents,
		submission.Status,
		submission.CreatedAt,
		submission.UpdatedAt,
	)
	return err
}

func (r *KYCSubmissionRepository) GetLatestByMerchant(ctx context.Context, merchantID string) (*models.KYCSubmission, error) {
	query := `
		SELECT id, merchant_id, business_type, business_name, cac_number, tin_number,
		       business_address, city, state, postal_code, incorporation_date,
			   business_category, director_name, director_bvn, director_phone, director_email,
			   documents, status, reviewer_id, review_notes, reviewed_at, created_at, updated_at
		FROM kyc_submissions
		WHERE merchant_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`
	var s models.KYCSubmission
	err := r.db.QueryRowContext(ctx, query, merchantID).Scan(
		&s.ID, &s.MerchantID, &s.BusinessType, &s.BusinessName, &s.CACNumber, &s.TINNumber,
		&s.BusinessAddress, &s.City, &s.State, &s.PostalCode, &s.IncorporationDate,
		&s.BusinessCategory, &s.DirectorName, &s.DirectorBVN, &s.DirectorPhone, &s.DirectorEmail,
		&s.Documents, &s.Status, &s.ReviewerID, &s.ReviewNotes, &s.ReviewedAt, &s.CreatedAt, &s.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *KYCSubmissionRepository) ListByStatus(ctx context.Context, status string, limit int) ([]*models.KYCSubmission, error) {
	if limit <= 0 {
		limit = 100
	}
	query := `
		SELECT id, merchant_id, business_type, business_name, status, created_at, updated_at
		FROM kyc_submissions
		WHERE status = $1
		ORDER BY created_at DESC
		LIMIT $2
	`
	rows, err := r.db.QueryContext(ctx, query, status, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*models.KYCSubmission
	for rows.Next() {
		var s models.KYCSubmission
		if err := rows.Scan(&s.ID, &s.MerchantID, &s.BusinessType, &s.BusinessName, &s.Status, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, &s)
	}
	return list, rows.Err()
}

func (r *KYCSubmissionRepository) UpdateStatus(ctx context.Context, id string, status string, reviewerID *string, notes *string) error {
	now := time.Now()
	query := `
		UPDATE kyc_submissions
		SET status = $2, reviewer_id = $3, review_notes = $4, reviewed_at = $5, updated_at = $6
		WHERE id = $1
	`
	res, err := r.db.ExecContext(ctx, query, id, status, reviewerID, notes, now, now)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("kyc submission not found")
	}
	return nil
}
