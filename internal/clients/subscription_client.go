package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type SubscriptionClient interface {
	CreateMerchantSubscription(ctx context.Context, merchantID int64, tierID int64) error
	GetMerchantSubscription(ctx context.Context, merchantID int64) (*SubscriptionResponse, error)
}

type subscriptionClient struct {
	baseURL    string
	httpClient *http.Client
}

type CreateSubscriptionRequest struct {
	MerchantID int64  `json:"merchant_id"`
	TierID     int64  `json:"tier_id"`
	IsTrial    bool   `json:"is_trial"`
	AutoRenew  bool   `json:"auto_renew"`
}

type SubscriptionResponse struct {
	ID                 int64     `json:"id"`
	MerchantID         int64     `json:"merchant_id"`
	TierID             int64     `json:"tier_id"`
	Status             string    `json:"status"`
	BillingCycle       string    `json:"billing_cycle"`
	CurrentPeriodStart time.Time `json:"current_period_start"`
	CurrentPeriodEnd   time.Time `json:"current_period_end"`
	AutoRenew          bool      `json:"auto_renew"`
}

func NewSubscriptionClient(baseURL string) SubscriptionClient {
	return &subscriptionClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *subscriptionClient) CreateMerchantSubscription(ctx context.Context, merchantID int64, tierID int64) error {
	req := CreateSubscriptionRequest{
		MerchantID: merchantID,
		TierID:     tierID,
		IsTrial:    false,
		AutoRenew:  true,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v1/subscription", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to call subscription service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("subscription service returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *subscriptionClient) GetMerchantSubscription(ctx context.Context, merchantID int64) (*SubscriptionResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/v1/subscription/%d", c.baseURL, merchantID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call subscription service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("subscription not found for merchant %d", merchantID)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("subscription service returned status %d: %s", resp.StatusCode, string(body))
	}

	var subscription SubscriptionResponse
	if err := json.NewDecoder(resp.Body).Decode(&subscription); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &subscription, nil
}
