package services

import (
	"context"
	"fmt"
	"time"

	"github.com/kodra-pay/merchant-service/internal/dto"
	"github.com/kodra-pay/merchant-service/internal/models"
	"github.com/kodra-pay/merchant-service/internal/repositories"
)

type PaymentLinkService struct {
	repo *repositories.PaymentLinkRepository
}

func NewPaymentLinkService(repo *repositories.PaymentLinkRepository) *PaymentLinkService {
	return &PaymentLinkService{repo: repo}
}

func (s *PaymentLinkService) CreatePaymentLink(ctx context.Context, req dto.CreatePaymentLinkRequest) (*dto.PaymentLinkResponse, error) {
	link := &models.PaymentLink{
		MerchantID:  req.MerchantID,
		Mode:        req.Mode,
		Amount:      req.Amount,
		Currency:    req.Currency,
		Description: req.Description,
		Status:      "active",
	}

	if err := s.repo.Create(ctx, link); err != nil {
		return nil, err
	}

	// Build checkout URL
	url := s.buildCheckoutURL(link)

	return &dto.PaymentLinkResponse{
		ID:          link.ID,
		MerchantID:  link.MerchantID,
		Mode:        link.Mode,
		Amount:      link.Amount,
		Currency:    link.Currency,
		Description: link.Description,
		Status:      link.Status,
		URL:         url,
		CreatedAt:   link.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   link.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

func (s *PaymentLinkService) GetPaymentLink(ctx context.Context, id string) (*dto.PaymentLinkResponse, error) {
	link, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if link == nil {
		return nil, nil
	}

	url := s.buildCheckoutURL(link)

	return &dto.PaymentLinkResponse{
		ID:          link.ID,
		MerchantID:  link.MerchantID,
		Mode:        link.Mode,
		Amount:      link.Amount,
		Currency:    link.Currency,
		Description: link.Description,
		Status:      link.Status,
		URL:         url,
		CreatedAt:   link.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   link.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

func (s *PaymentLinkService) ListPaymentLinks(ctx context.Context, merchantID string, limit int) (*dto.ListPaymentLinksResponse, error) {
	links, err := s.repo.GetByMerchantID(ctx, merchantID, limit)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.PaymentLinkResponse, 0, len(links))
	for _, link := range links {
		url := s.buildCheckoutURL(&link)
		responses = append(responses, dto.PaymentLinkResponse{
			ID:          link.ID,
			MerchantID:  link.MerchantID,
			Mode:        link.Mode,
			Amount:      link.Amount,
			Currency:    link.Currency,
			Description: link.Description,
			Status:      link.Status,
			URL:         url,
			CreatedAt:   link.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   link.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return &dto.ListPaymentLinksResponse{
		Links: responses,
		Total: len(responses),
	}, nil
}

func (s *PaymentLinkService) buildCheckoutURL(link *models.PaymentLink) string {
	// Build the checkout URL based on the payment link
	baseURL := "http://localhost:5174/merchant/checkout"
	url := fmt.Sprintf("%s?ref=%s&currency=%s&mode=%s&merchant_id=%s",
		baseURL, link.ID, link.Currency, link.Mode, link.MerchantID)

	if link.Description != "" {
		url += fmt.Sprintf("&description=%s", link.Description)
	}

	if link.Mode == "fixed" && link.Amount != nil {
		url += fmt.Sprintf("&amount=%d", *link.Amount)
	}

	return url
}

// Helper function to get current timestamp in milliseconds
func getCurrentTimestampMillis() int64 {
	return time.Now().UnixMilli()
}
