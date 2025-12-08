package dto

type MerchantBalanceResponse struct {
	MerchantID       int     `json:"merchant_id"`
	Currency         string  `json:"currency"`
	PendingBalance   float64 `json:"pending_balance"`   // In currency units (e.g., NGN)
	AvailableBalance float64 `json:"available_balance"` // In currency units (e.g., NGN)
	TotalVolume      float64 `json:"total_volume"`      // In currency units (e.g., NGN)
}
