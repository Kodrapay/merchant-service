package dto

import "time"

// WalletCreateRequest represents the payload for provisioning a wallet
type WalletCreateRequest struct {
	UserID   string `json:"user_id"`
	Currency string `json:"currency"`
}

// WalletResponse mirrors the wallet-ledger service response
type WalletResponse struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Currency  string    `json:"currency"`
	Balance   int64     `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
