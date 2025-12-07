package dto

type KYCSubmissionRequest struct {
	MerchantID        int               `json:"merchant_id"`
	BusinessType      string            `json:"business_type"` // "registered" or "startup"
	BusinessName      string            `json:"business_name"`
	CACNumber         string            `json:"cac_number,omitempty"`
	TINNumber         string            `json:"tin_number,omitempty"`
	BusinessAddress   string            `json:"business_address"`
	City              string            `json:"city"`
	State             string            `json:"state"`
	PostalCode        string            `json:"postal_code,omitempty"`
	IncorporationDate string            `json:"incorporation_date,omitempty"`
	BusinessCategory  string            `json:"business_category"`
	DirectorName      string            `json:"director_name"`
	DirectorBVN       string            `json:"director_bvn"`
	DirectorPhone     string            `json:"director_phone"`
	DirectorEmail     string            `json:"director_email"`
	Documents         map[string]string `json:"documents"` // document_type -> file_path/url
}

type KYCSubmissionResponse struct {
	SubmissionID int    `json:"submission_id"`
	Status       string `json:"status"`
	Message      string `json:"message"`
}

type KYCStatusUpdateRequest struct {
	MerchantID  int    `json:"merchant_id"`
	Status      string `json:"status"` // "approved" or "rejected"
	ReviewerID  int    `json:"reviewer_id"`
	ReviewNotes string `json:"review_notes,omitempty"`
}

type KYCStatusResponse struct {
	MerchantID  int    `json:"merchant_id"`
	Status      string `json:"status"`
	SubmittedAt string `json:"submitted_at,omitempty"`
	ReviewedAt  string `json:"reviewed_at,omitempty"`
	ReviewerID  *int   `json:"reviewer_id,omitempty"`
	ReviewNotes string `json:"review_notes,omitempty"`
}
