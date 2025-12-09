package services

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
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

func (s *PaymentLinkService) DeletePaymentLink(ctx context.Context, id, merchantID int) error {
	if id == 0 { // int check
		return fmt.Errorf("id is required")
	}
	if err := s.repo.Delete(ctx, id, merchantID); err != nil {
		if errors.Is(err, repositories.ErrPaymentLinkNotFound) || errors.Is(err, sql.ErrNoRows) {
			return repositories.ErrPaymentLinkNotFound
		}
		return err
	}
	return nil
}

func (s *PaymentLinkService) CreatePaymentLink(ctx context.Context, req dto.CreatePaymentLinkRequest) (*dto.PaymentLinkResponse, error) {
	var amountKobo *int64
	if req.Amount != nil {
		val := int64(math.Round(*req.Amount * 100))
		amountKobo = &val
	}

	link := &models.PaymentLink{
		MerchantID:  req.MerchantID,
		Mode:        req.Mode,
		Amount:      amountKobo,
		Currency:    req.Currency,
		Description: req.Description,
		Status:      "active",
	}

	// Generate signature for tampering detection
	link.Signature = generatePaymentLinkSignature(link)

	if err := s.repo.Create(ctx, link); err != nil {
		return nil, err
	}

	// Build checkout URL
	url := s.buildCheckoutURL(link)

	return &dto.PaymentLinkResponse{
		ID:          link.ID,
		MerchantID:  link.MerchantID,
		Mode:        link.Mode,
		Amount:      toCurrencyAmount(link.Amount),
		Currency:    link.Currency,
		Description: link.Description,
		Status:      link.Status,
		URL:         url,
		CreatedAt:   link.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   link.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

func (s *PaymentLinkService) GetPaymentLink(ctx context.Context, id int) (*dto.PaymentLinkResponse, error) {
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
		Amount:      toCurrencyAmount(link.Amount),
		Currency:    link.Currency,
		Description: link.Description,
		Status:      link.Status,
		URL:         url,
		CreatedAt:   link.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   link.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

func (s *PaymentLinkService) ListPaymentLinks(ctx context.Context, merchantID int, limit int) (*dto.ListPaymentLinksResponse, error) {
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
			Amount:      toCurrencyAmount(link.Amount),
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
	url := fmt.Sprintf("%s?ref=%d&currency=%s&mode=%s&merchant_id=%d",
		baseURL, link.ID, link.Currency, link.Mode, link.MerchantID)

	if link.Description != "" {
		url += fmt.Sprintf("&description=%s", link.Description)
	}

	if link.Mode == "fixed" && link.Amount != nil {
		url += fmt.Sprintf("&amount=%.2f", float64(*link.Amount)/100)
	}

	return url
}

func toCurrencyAmount(amountKobo *int64) *float64 {
	if amountKobo == nil {
		return nil
	}
	val := float64(*amountKobo) / 100
	return &val
}

// Helper function to get current timestamp in milliseconds
func getCurrentTimestampMillis() int64 {
	return time.Now().UnixMilli()
}

// generatePaymentLinkSignature creates an HMAC-SHA256 signature of the payment link parameters
func generatePaymentLinkSignature(link *models.PaymentLink) string {
	secret := os.Getenv("PAYMENT_LINK_SECRET")
	if secret == "" {
		secret = "kodrapay-default-secret-change-in-production"
	}

	// Create canonical string from link parameters
	amountStr := "null"
	if link.Amount != nil {
		amountStr = strconv.FormatInt(*link.Amount, 10)
	}

	data := fmt.Sprintf("%d|%s|%s|%s|%s",
		link.MerchantID,
		link.Mode,
		amountStr,
		link.Currency,
		link.Description,
	)

	// Generate HMAC-SHA256 signature
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// VerifyPaymentLinkSignature verifies that a payment link's parameters haven't been tampered with
func VerifyPaymentLinkSignature(link *models.PaymentLink) bool {
	if link.Signature == "" {
		// Old links without signature - allow them but log warning
		fmt.Printf("Warning: Payment link %d has no signature\n", link.ID)
		return true
	}

	expectedSignature := generatePaymentLinkSignature(link)
	return hmac.Equal([]byte(link.Signature), []byte(expectedSignature))
}
