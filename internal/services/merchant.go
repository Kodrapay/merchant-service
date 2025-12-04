package services

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/kodra-pay/merchant-service/internal/dto"
	"github.com/kodra-pay/merchant-service/internal/models"
	"github.com/kodra-pay/merchant-service/internal/repositories"
)

type MerchantService struct {
	repo       *repositories.MerchantRepository
	apiKeyRepo *repositories.APIKeyRepository
}

func NewMerchantService(repo *repositories.MerchantRepository, apiKeyRepo *repositories.APIKeyRepository) *MerchantService {
	return &MerchantService{repo: repo, apiKeyRepo: apiKeyRepo}
}

func (s *MerchantService) List(ctx context.Context) []dto.MerchantResponse {
	merchants, err := s.repo.List(ctx, "", 100, 0)
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

func (s *MerchantService) Create(ctx context.Context, req dto.MerchantCreateRequest) dto.MerchantCreateResponse {
	merchant := &models.Merchant{
		ID:           uuid.NewString(),
		Name:         req.Name,
		Email:        req.Email,
		BusinessName: req.BusinessName,
		Country:      req.Country,
		Status:       models.MerchantStatusInactive,
		KYCStatus:    models.KYCStatusNotStarted,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := s.repo.Create(ctx, merchant)
	if err != nil {
		// Log error in production
		return dto.MerchantCreateResponse{ID: ""}
	}

	return dto.MerchantCreateResponse{ID: merchant.ID}
}

func (s *MerchantService) Get(ctx context.Context, id string) dto.MerchantResponse {
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

// GetMerchant returns the merchant model (for internal use)
func (s *MerchantService) GetMerchant(ctx context.Context, id string) (*models.Merchant, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *MerchantService) UpdateStatus(ctx context.Context, id string, req dto.MerchantStatusUpdateRequest) map[string]string {
	status := models.MerchantStatus(req.Status)
	err := s.repo.UpdateStatus(ctx, id, status)
	if err != nil {
		return map[string]string{"id": id, "status": "error", "message": err.Error()}
	}
	return map[string]string{"id": id, "status": req.Status}
}

func (s *MerchantService) ListAPIKeys(ctx context.Context, id string) []dto.APIKeyResponse {
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

func (s *MerchantService) RotateAPIKey(ctx context.Context, id string) dto.APIKeyResponse {
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
