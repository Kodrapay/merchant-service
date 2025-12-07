package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/kodra-pay/merchant-service/internal/dto"
)

// ErrWalletNotFound is returned when the wallet service replies with 404.
var ErrWalletNotFound = errors.New("wallet not found")

// WalletLedgerClient defines the interaction surface we need for wallets.
type WalletLedgerClient interface {
	GetWalletByUserIDAndCurrency(ctx context.Context, userID, currency string) (*dto.WalletResponse, error)
	CreateWallet(ctx context.Context, req dto.WalletCreateRequest) (*dto.WalletResponse, error)
}

type httpWalletLedgerClient struct {
	baseURL string
	client  *http.Client
}

// NewHTTPWalletLedgerClient creates a new client with sane timeouts.
func NewHTTPWalletLedgerClient(baseURL string) WalletLedgerClient {
	return &httpWalletLedgerClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *httpWalletLedgerClient) GetWalletByUserIDAndCurrency(ctx context.Context, userID, currency string) (*dto.WalletResponse, error) {
	query := fmt.Sprintf("%s/wallets?user_id=%s&currency=%s", c.baseURL, url.QueryEscape(userID), url.QueryEscape(currency))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, query, nil)
	if err != nil {
		return nil, fmt.Errorf("create wallet lookup request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("wallet-ledger request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrWalletNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("wallet-ledger returned status %d", resp.StatusCode)
	}

	var wallet dto.WalletResponse
	if err := json.NewDecoder(resp.Body).Decode(&wallet); err != nil {
		return nil, fmt.Errorf("decode wallet response: %w", err)
	}
	return &wallet, nil
}

func (c *httpWalletLedgerClient) CreateWallet(ctx context.Context, req dto.WalletCreateRequest) (*dto.WalletResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal wallet create request: %w", err)
	}

	url := fmt.Sprintf("%s/wallets", c.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("create wallet request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("wallet-ledger request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("wallet-ledger returned status %d for create", resp.StatusCode)
	}

	var wallet dto.WalletResponse
	if err := json.NewDecoder(resp.Body).Decode(&wallet); err != nil {
		return nil, fmt.Errorf("decode wallet create response: %w", err)
	}
	return &wallet, nil
}
