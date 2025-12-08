package models

import "time"

type MerchantBalance struct {
	ID               int       `json:"id" db:"id"`
	MerchantID       int       `json:"merchant_id" db:"merchant_id"`
	Currency         string    `json:"currency" db:"currency"`
	PendingBalance   int64     `json:"pending_balance" db:"pending_balance"`     // Unsettled funds (kobo)
	AvailableBalance int64     `json:"available_balance" db:"available_balance"` // Settled funds (kobo)
	TotalVolume      int64     `json:"total_volume" db:"total_volume"`           // Total processed (kobo)
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}
