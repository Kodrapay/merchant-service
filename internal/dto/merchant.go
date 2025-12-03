package dto

type MerchantCreateRequest struct {
	Name         string `json:"name"`
	Email        string `json:"email"`
	BusinessName string `json:"business_name"`
	Country      string `json:"country"`
}

type MerchantCreateResponse struct {
	ID string `json:"id"`
}

type MerchantStatusUpdateRequest struct {
	Status string `json:"status"`
}

type MerchantResponse struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	BusinessName string `json:"business_name"`
	Status       string `json:"status"`
	Country      string `json:"country"`
}

type APIKeyResponse struct {
	KeyID string `json:"key_id"`
	Key   string `json:"key"`
	Type  string `json:"type"`
}
