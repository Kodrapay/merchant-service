package dto

type CreatePaymentLinkRequest struct {
	MerchantID  string `json:"merchant_id"`
	Mode        string `json:"mode"`
	Amount      *int64 `json:"amount,omitempty"`
	Currency    string `json:"currency"`
	Description string `json:"description"`
}

type PaymentLinkResponse struct {
	ID          string `json:"id"`
	MerchantID  string `json:"merchant_id"`
	Mode        string `json:"mode"`
	Amount      *int64 `json:"amount,omitempty"`
	Currency    string `json:"currency"`
	Description string `json:"description"`
	Status      string `json:"status"`
	URL         string `json:"url"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type ListPaymentLinksResponse struct {
	Links []PaymentLinkResponse `json:"links"`
	Total int                   `json:"total"`
}
