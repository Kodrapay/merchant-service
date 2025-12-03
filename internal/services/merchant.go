package services

import (
	"context"

	"github.com/google/uuid"

	"github.com/kodra-pay/merchant-service/internal/dto"
	"github.com/kodra-pay/merchant-service/internal/repositories"
)

type MerchantService struct {
	repo *repositories.MerchantRepository
}

func NewMerchantService(repo *repositories.MerchantRepository) *MerchantService {
	return &MerchantService{repo: repo}
}

func (s *MerchantService) List(_ context.Context) []dto.MerchantResponse {
	return []dto.MerchantResponse{}
}

func (s *MerchantService) Create(_ context.Context, req dto.MerchantCreateRequest) dto.MerchantCreateResponse {
	return dto.MerchantCreateResponse{ID: uuid.NewString()}
}

func (s *MerchantService) Get(_ context.Context, id string) dto.MerchantResponse {
	return dto.MerchantResponse{
		ID:           id,
		Name:         "stub",
		Email:        "merchant@example.com",
		BusinessName: "Stub Business",
		Status:       "pending",
		Country:      "NG",
	}
}

func (s *MerchantService) UpdateStatus(_ context.Context, id string, req dto.MerchantStatusUpdateRequest) map[string]string {
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
