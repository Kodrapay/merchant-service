package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/kodra-pay/merchant-service/internal/clients"
	"github.com/kodra-pay/merchant-service/internal/dto"
	"github.com/kodra-pay/merchant-service/internal/models"
	"github.com/kodra-pay/merchant-service/internal/repositories"
)

type MerchantService struct {
	repo               *repositories.MerchantRepository
	apiKeyRepo         *repositories.APIKeyRepository
	settlementRepo     *repositories.SettlementConfigRepository
	walletLedgerClient clients.WalletLedgerClient
	subscriptionClient clients.SubscriptionClient
}

func NewMerchantService(repo *repositories.MerchantRepository, apiKeyRepo *repositories.APIKeyRepository, settlementRepo *repositories.SettlementConfigRepository, walletLedgerClient clients.WalletLedgerClient, subscriptionClient clients.SubscriptionClient) *MerchantService {
	return &MerchantService{
		repo:               repo,
		apiKeyRepo:         apiKeyRepo,
		settlementRepo:     settlementRepo,
		walletLedgerClient: walletLedgerClient,
		subscriptionClient: subscriptionClient,
	}
}

func (s *MerchantService) List(ctx context.Context) []dto.MerchantResponse {
	merchants, err := s.repo.List(ctx, "", 100, 0)
	if err != nil {
		log.Printf("ERROR: MerchantService.List - error from repository: %v", err)
		return []dto.MerchantResponse{}
	}
	log.Printf("DEBUG: MerchantService.List - merchants retrieved: %d", len(merchants))

	responses := make([]dto.MerchantResponse, len(merchants))
	for i, m := range merchants {
		responses[i] = dto.MerchantResponse{
			ID:           m.ID,
			Name:         m.Name,
			Email:        m.Email,
			BusinessName: m.BusinessName,
			Status:       string(m.Status),
			KYCStatus:    string(m.KYCStatus),
			Country:      m.Country,
			CanTransact:  m.CanTransact(),
		}
	}

	return responses
}

func (s *MerchantService) Create(ctx context.Context, req dto.MerchantCreateRequest) dto.MerchantCreateResponse {
	merchant := &models.Merchant{
		Name:         req.Name,
		Email:        req.Email,
		BusinessName: req.BusinessName,
		Country:      req.Country,
		Status:       models.MerchantStatusInactive, // Set initial status
		KYCStatus:    models.KYCStatusNotStarted,    // Set initial KYC status
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := s.repo.Create(ctx, merchant)
	if err != nil {
		// Log error in production
		return dto.MerchantCreateResponse{ID: 0}
	}

	// After creating the merchant, create a wallet for them
	if err := s.ensureMerchantWallet(ctx, merchant.ID, "NGN"); err != nil {
		log.Printf("Failed to provision wallet for merchant %d: %v", merchant.ID, err)
		// Depending on business logic, you might want to handle this error differently
	} else {
		log.Printf("Successfully provisioned wallet for merchant %d", merchant.ID)
	}

	// Provision default settlement config (idempotent)
	if err := s.ensureSettlementConfig(ctx, merchant.ID); err != nil {
		log.Printf("Failed to provision default settlement config for merchant %d: %v", merchant.ID, err)
	}

	// Create subscription with Starter tier (tier_id = 1)
	if s.subscriptionClient != nil {
		if err := s.subscriptionClient.CreateMerchantSubscription(ctx, int64(merchant.ID), 1); err != nil {
			log.Printf("Failed to create subscription for merchant %d: %v", merchant.ID, err)
			// Don't fail merchant creation if subscription fails
		} else {
			log.Printf("Successfully created subscription for merchant %d", merchant.ID)
		}
	}

	return dto.MerchantCreateResponse{ID: merchant.ID}
}

func (s *MerchantService) Get(ctx context.Context, id int) dto.MerchantResponse {
	merchant, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return dto.MerchantResponse{}
	}

	return dto.MerchantResponse{
		ID:           merchant.ID,
		Name:         merchant.Name,
		Email:        merchant.Email,
		BusinessName: merchant.BusinessName,
		Status:       string(merchant.Status),
		KYCStatus:    string(merchant.KYCStatus),
		Country:      merchant.Country,
		CanTransact:  merchant.CanTransact(),
	}
}

// GetAny returns the first merchant (fallback when id is not provided)
func (s *MerchantService) GetAny(ctx context.Context) dto.MerchantResponse {
	merchants, err := s.repo.List(ctx, "", 1, 0)
	if err != nil || len(merchants) == 0 {
		return dto.MerchantResponse{}
	}
	m := merchants[0]
	return dto.MerchantResponse{
		ID:           m.ID,
		Name:         m.Name,
		Email:        m.Email,
		BusinessName: m.BusinessName,
		Status:       string(m.Status),
		KYCStatus:    string(m.KYCStatus),
		Country:      m.Country,
		CanTransact:  m.CanTransact(),
	}
}

// GetByEmail is a helper to retrieve by email if needed in other flows
func (s *MerchantService) GetByEmail(ctx context.Context, email string) dto.MerchantResponse {
	merchant, err := s.repo.GetByEmail(ctx, email)
	if err != nil || merchant == nil {
		return dto.MerchantResponse{}
	}
	return dto.MerchantResponse{
		ID:           merchant.ID,
		Name:         merchant.Name,
		Email:        merchant.Email,
		BusinessName: merchant.BusinessName,
		Status:       string(merchant.Status),
		KYCStatus:    string(merchant.KYCStatus),
		Country:      merchant.Country,
		CanTransact:  merchant.CanTransact(),
	}
}

// GetMerchant returns the merchant model (for internal use)
func (s *MerchantService) GetMerchant(ctx context.Context, id int) (*models.Merchant, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *MerchantService) ListByKYCStatus(ctx context.Context, kycStatus models.KYCStatus, limit, offset int) []dto.MerchantResponse {
	merchants, err := s.repo.ListByKYCStatus(ctx, kycStatus, limit, offset)
	if err != nil {
		return []dto.MerchantResponse{}
	}

	responses := make([]dto.MerchantResponse, len(merchants))
	for i, m := range merchants {
		responses[i] = dto.MerchantResponse{
			ID:           m.ID,
			Name:         m.Name,
			Email:        m.Email,
			BusinessName: m.BusinessName,
			Status:       string(m.Status),
			KYCStatus:    string(m.KYCStatus),
			Country:      m.Country,
			CanTransact:  m.CanTransact(),
		}
	}

	return responses
}

func (s *MerchantService) UpdateKYCStatus(ctx context.Context, id int, req dto.MerchantKYCStatusUpdateRequest) map[string]interface{} {
	statusValue := strings.ToLower(req.KYCStatus)
	if statusValue == "completed" {
		statusValue = string(models.KYCStatusApproved)
	}

	kycStatus := models.KYCStatus(statusValue)

	err := s.repo.UpdateKYCStatus(ctx, id, kycStatus)
	if err != nil {
		return map[string]interface{}{"id": id, "kyc_status": "error", "message": err.Error()}
	}

	resp := map[string]interface{}{"id": id, "kyc_status": string(kycStatus)}

	// Provision a wallet for approved merchants (idempotent check against wallet-ledger service).
	if kycStatus == models.KYCStatusApproved {
		if err := s.ensureMerchantWallet(ctx, id, "NGN"); err != nil {
			log.Printf("Failed to provision wallet for merchant %d: %v", id, err)
			resp["wallet_status"] = "error"
			resp["wallet_error"] = err.Error()
		} else {
			resp["wallet_status"] = "created"
		}

		// Provision default settlement config (idempotent)
		if err := s.ensureSettlementConfig(ctx, id); err != nil {
			log.Printf("Failed to provision settlement config for merchant %d: %v", id, err)
			resp["settlement_config"] = "error"
			resp["settlement_error"] = err.Error()
		} else {
			resp["settlement_config"] = "created"
		}
	}

	return resp
}

// ensureSettlementConfig creates a default settlement config if missing (idempotent via GetByMerchantID)
func (s *MerchantService) ensureSettlementConfig(ctx context.Context, merchantID int) error {
	if s.settlementRepo == nil {
		return fmt.Errorf("settlement repository not configured")
	}
	_, err := s.settlementRepo.GetByMerchantID(ctx, merchantID)
	return err
}

func (s *MerchantService) UpdateStatus(ctx context.Context, id int, req dto.MerchantStatusUpdateRequest) map[string]interface{} {
	status := models.MerchantStatus(req.Status)
	err := s.repo.UpdateStatus(ctx, id, status)
	if err != nil {
		return map[string]interface{}{"id": id, "status": "error", "message": err.Error()}
	}
	return map[string]interface{}{"id": id, "status": req.Status}
}

func (s *MerchantService) ListAPIKeys(ctx context.Context, id int) []dto.APIKeyResponse {
	keys, err := s.apiKeyRepo.ListByMerchantID(ctx, id)
	if err != nil {
		return []dto.APIKeyResponse{}
	}

	// If no keys exist, create default test keys
	if len(keys) == 0 {
		// Create public test key
		pubKey, pubFullKey, _ := models.GenerateAPIKey(id, models.APIKeyTypePublic, models.EnvironmentTest)
		if pubKey != nil {
			_ = s.apiKeyRepo.Create(ctx, pubKey)
			keys = append(keys, pubKey)
		}

		// Create secret test key
		secKey, secFullKey, _ := models.GenerateAPIKey(id, models.APIKeyTypeSecret, models.EnvironmentTest)
		if secKey != nil {
			_ = s.apiKeyRepo.Create(ctx, secKey)
			keys = append(keys, secKey)
		}

		// Return with full keys on first creation
		responses := make([]dto.APIKeyResponse, 0, len(keys))
		if pubKey != nil {
			responses = append(responses, dto.APIKeyResponse{
				KeyID:       pubKey.ID,
				Key:         pubFullKey,
				KeyPrefix:   pubKey.KeyPrefix,
				Type:        string(pubKey.KeyType),
				Environment: string(pubKey.Environment),
				CreatedAt:   pubKey.CreatedAt.Format(time.RFC3339),
			})
		}
		if secKey != nil {
			responses = append(responses, dto.APIKeyResponse{
				KeyID:       secKey.ID,
				Key:         secFullKey,
				KeyPrefix:   secKey.KeyPrefix,
				Type:        string(secKey.KeyType),
				Environment: string(secKey.Environment),
				CreatedAt:   secKey.CreatedAt.Format(time.RFC3339),
			})
		}
		return responses
	}

	// Return existing keys without full key value
	responses := make([]dto.APIKeyResponse, len(keys))
	for i, key := range keys {
		responses[i] = dto.APIKeyResponse{
			KeyID:       key.ID,
			KeyPrefix:   key.KeyPrefix,
			Type:        string(key.KeyType),
			Environment: string(key.Environment),
			CreatedAt:   key.CreatedAt.Format(time.RFC3339),
		}
	}
	return responses
}

// ListByKYCStatuses returns a list of merchants filtered by multiple KYC statuses
func (s *MerchantService) ListByKYCStatuses(ctx context.Context, kycStatuses []models.KYCStatus, limit, offset int) []dto.MerchantResponse {
	log.Printf("DEBUG: MerchantService.ListByKYCStatuses called with kycStatuses: %v", kycStatuses)
	merchants, err := s.repo.ListByKYCStatuses(ctx, kycStatuses, limit, offset)
	if err != nil {
		log.Printf("ERROR: MerchantService.ListByKYCStatuses - error from repository: %v", err)
		return []dto.MerchantResponse{}
	}
	log.Printf("DEBUG: MerchantService.ListByKYCStatuses - kycStatuses: %v, merchants retrieved: %d", kycStatuses, len(merchants))

	responses := make([]dto.MerchantResponse, len(merchants))
	for i, m := range merchants {
		responses[i] = dto.MerchantResponse{
			ID:           m.ID,
			Name:         m.Name,
			Email:        m.Email,
			BusinessName: m.BusinessName,
			Status:       string(m.Status),
			KYCStatus:    string(m.KYCStatus),
			Country:      m.Country,
			CanTransact:  m.CanTransact(),
		}
	}

	return responses
}

func (s *MerchantService) RotateAPIKey(ctx context.Context, id int) dto.APIKeyResponse {
	// Deactivate old secret test key
	_ = s.apiKeyRepo.DeactivateByMerchantAndType(ctx, id, models.APIKeyTypeSecret, models.EnvironmentTest)

	// Generate new secret test key
	newKey, fullKey, err := models.GenerateAPIKey(id, models.APIKeyTypeSecret, models.EnvironmentTest)
	if err != nil {
		return dto.APIKeyResponse{}
	}

	if err := s.apiKeyRepo.Create(ctx, newKey); err != nil {
		return dto.APIKeyResponse{}
	}

	return dto.APIKeyResponse{
		KeyID:       newKey.ID,
		Key:         fullKey,
		KeyPrefix:   newKey.KeyPrefix,
		Type:        string(newKey.KeyType),
		Environment: string(newKey.Environment),
		CreatedAt:   newKey.CreatedAt.Format(time.RFC3339),
	}
}

// ensureMerchantWallet checks for an existing wallet and creates one if missing.
func (s *MerchantService) ensureMerchantWallet(ctx context.Context, merchantID int, currency string) error {
	if s.walletLedgerClient == nil {
		return fmt.Errorf("wallet-ledger client not configured")
	}

	wallet, err := s.walletLedgerClient.GetWalletByUserIDAndCurrency(ctx, merchantID, currency)
	if err != nil && !errors.Is(err, clients.ErrWalletNotFound) {
		return err
	}
	if wallet != nil {
		return nil
	}

	_, err = s.walletLedgerClient.CreateWallet(ctx, dto.WalletCreateRequest{
		UserID:   merchantID,
		Currency: currency,
	})
	return err
}
