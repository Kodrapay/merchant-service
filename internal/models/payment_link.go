package models

import "time"

type PaymentLink struct {
	ID          string     `json:"id"`
	MerchantID  string     `json:"merchant_id"`
	Mode        string     `json:"mode"`
	Amount      *int64     `json:"amount,omitempty"`
	Currency    string     `json:"currency"`
	Description string     `json:"description,omitempty"`
	Status      string     `json:"status"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
