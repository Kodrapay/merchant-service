package models

import "time"

type KYCSubmission struct {
	ID                int               `json:"id"`
	MerchantID        int               `json:"merchant_id"`
	BusinessType      string            `json:"business_type"`
	BusinessName      string            `json:"business_name"`
	CACNumber         string            `json:"cac_number,omitempty"`
	TINNumber         string            `json:"tin_number,omitempty"`
	BusinessAddress   string            `json:"business_address"`
	City              string            `json:"city"`
	State             string            `json:"state"`
	PostalCode        string            `json:"postal_code,omitempty"`
	IncorporationDate *time.Time        `json:"incorporation_date,omitempty"`
	BusinessCategory  string            `json:"business_category"`
	DirectorName      string            `json:"director_name"`
	DirectorBVN       string            `json:"director_bvn"`
	DirectorPhone     string            `json:"director_phone"`
	DirectorEmail     string            `json:"director_email"`
	Documents         map[string]string `json:"documents"`
	Status            string            `json:"status"`
	ReviewerID        *int              `json:"reviewer_id,omitempty"`
	ReviewNotes       *string           `json:"review_notes,omitempty"`
	ReviewedAt        *time.Time        `json:"reviewed_at,omitempty"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
}
