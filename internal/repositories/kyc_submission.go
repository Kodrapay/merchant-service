package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

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
	submission.Status = "pending"
	submission.CreatedAt = now
	submission.UpdatedAt = now

	var documents []byte
	if submission.Documents != nil {
		var err error
		documents, err = json.Marshal(submission.Documents)
		if err != nil {
			return fmt.Errorf("failed to marshal documents: %w", err)
		}
	}

	query := `
		INSERT INTO kyc_submissions (
			merchant_id, business_type, business_name, cac_number, tin_number,
			business_address, city, state, postal_code, incorporation_date,
			business_category, director_name, director_bvn, director_phone, director_email,
			documents, status, created_at, updated_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19
		)
		RETURNING id
	`
	var id int
	err := r.db.QueryRowContext(ctx, query,
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
	).Scan(&id)

	if err == nil {
		submission.ID = id
	}
	return err
}

func (r *KYCSubmissionRepository) GetLatestByMerchant(ctx context.Context, merchantID int) (*models.KYCSubmission, error) {
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
	var documents []byte
	var reviewerID sql.NullInt32

	err := r.db.QueryRowContext(ctx, query, merchantID).Scan(
		&s.ID, &s.MerchantID, &s.BusinessType, &s.BusinessName, &s.CACNumber, &s.TINNumber,
		&s.BusinessAddress, &s.City, &s.State, &s.PostalCode, &s.IncorporationDate,
		&s.BusinessCategory, &s.DirectorName, &s.DirectorBVN, &s.DirectorPhone, &s.DirectorEmail,
		&documents, &s.Status, &reviewerID, &s.ReviewNotes, &s.ReviewedAt, &s.CreatedAt, &s.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if len(documents) > 0 {
		if err := json.Unmarshal(documents, &s.Documents); err != nil {
			return nil, fmt.Errorf("failed to unmarshal documents: %w", err)
		}
	}

	if reviewerID.Valid {
		val := int(reviewerID.Int32)
		s.ReviewerID = &val
	} else {
		s.ReviewerID = nil
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

func (r *KYCSubmissionRepository) UpdateStatus(ctx context.Context, id int, status string, reviewerID *int, notes *string) error {
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
