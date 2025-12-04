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
	repo *repositories.MerchantRepository
}

func NewMerchantService(repo *repositories.MerchantRepository) *MerchantService {
	return &MerchantService{repo: repo}
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

func (s *MerchantService) ListAPIKeys(_ context.Context, id string) []dto.APIKeyResponse {
	return []dto.APIKeyResponse{
		{KeyID: "pub_" + uuid.NewString(), Key: "pk_test_stub", Type: "public"},
		{KeyID: "sec_" + uuid.NewString(), Key: "sk_test_stub", Type: "secret"},
	}
}

func (s *MerchantService) RotateAPIKey(_ context.Context, id string) dto.APIKeyResponse {
	return dto.APIKeyResponse{KeyID: "rotated_" + uuid.NewString(), Key: "sk_test_rotated", Type: "secret"}
}
