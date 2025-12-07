package dto

type MerchantCreateRequest struct {
	Name         string `json:"name"`
	Email        string `json:"email"`
	BusinessName string `json:"business_name"`
	Country      string `json:"country"`
}

type MerchantCreateResponse struct {
	ID int `json:"id"`
}

type MerchantStatusUpdateRequest struct {
	Status string `json:"status"`
}

type MerchantKYCStatusUpdateRequest struct {
	KYCStatus string `json:"kyc_status"`
}

type MerchantResponse struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	BusinessName string `json:"business_name"`
	Status       string `json:"status"`
	KYCStatus    string `json:"kyc_status"`
	Country      string `json:"country"`
	CanTransact  bool   `json:"can_transact"`
}

type APIKeyResponse struct {
	KeyID       int    `json:"key_id"`
	Key         string `json:"key,omitempty"` // Only returned when creating/rotating
	KeyPrefix   string `json:"key_prefix"`
	Type        string `json:"type"`
	Environment string `json:"environment"`
	CreatedAt   string `json:"created_at"`
}
