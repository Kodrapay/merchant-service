package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kodra-pay/merchant-service/internal/dto"
	"github.com/kodra-pay/merchant-service/internal/models"
	"github.com/kodra-pay/merchant-service/internal/repositories"
)

type KYCService struct {
	merchantRepo *repositories.MerchantRepository
	kycRepo      *repositories.KYCSubmissionRepository
}

func NewKYCService(merchantRepo *repositories.MerchantRepository, kycRepo *repositories.KYCSubmissionRepository) *KYCService {
	return &KYCService{
		merchantRepo: merchantRepo,
		kycRepo:      kycRepo,
	}
}

func (s *KYCService) Submit(ctx context.Context, req dto.KYCSubmissionRequest) (*dto.KYCSubmissionResponse, error) {
	if req.MerchantID == 0 { // int check
		return nil, fmt.Errorf("merchant_id is required")
	}

	businessType := strings.ToLower(strings.TrimSpace(req.BusinessType))
	if businessType == "" {
		businessType = "registered"
	}
	if businessType != "registered" && businessType != "startup" && businessType != "small_business" {
		return nil, fmt.Errorf("business_type must be registered, startup, or small_business")
	}

	merchant, err := s.merchantRepo.GetByID(ctx, req.MerchantID) // req.MerchantID is int
	if err != nil {
		return nil, fmt.Errorf("merchant not found")
	}

	submission := &models.KYCSubmission{
		MerchantID:       merchant.ID, // merchant.ID is int
		BusinessType:     businessType,
		BusinessName:     req.BusinessName,
		CACNumber:        req.CACNumber,
		TINNumber:        req.TINNumber,
		BusinessAddress:  req.BusinessAddress,
		City:             req.City,
		State:            req.State,
		PostalCode:       req.PostalCode,
		BusinessCategory: req.BusinessCategory,
		DirectorName:     req.DirectorName,
		DirectorBVN:      req.DirectorBVN,
		DirectorPhone:    req.DirectorPhone,
		DirectorEmail:    req.DirectorEmail,
		Documents:        req.Documents,
	}

	if req.IncorporationDate != "" {
		if parsed, err := time.Parse("2006-01-02", req.IncorporationDate); err == nil {
			submission.IncorporationDate = &parsed
		}
	}

	if err := s.kycRepo.Create(ctx, submission); err != nil {
		return nil, err
	}

	// move merchant into pending KYC so admin can review
	_ = s.merchantRepo.UpdateKYCStatus(ctx, merchant.ID, models.KYCStatusPending)

	return &dto.KYCSubmissionResponse{
		SubmissionID: submission.ID, // submission.ID is int
		Status:       "pending",
		Message:      "KYC submission received and is under review",
	}, nil
}

func (s *KYCService) GetLatest(ctx context.Context, merchantID int) (*dto.KYCStatusResponse, error) {
	submission, err := s.kycRepo.GetLatestByMerchant(ctx, merchantID) // merchantID is int
	if err != nil {
		return nil, err
	}
	if submission == nil {
		return nil, nil
	}

	return &dto.KYCStatusResponse{
		MerchantID:  submission.MerchantID, // int
		Status:      submission.Status,
		SubmittedAt: submission.CreatedAt.Format(time.RFC3339),
		ReviewedAt:  timePtrToString(submission.ReviewedAt),
		ReviewerID:  submission.ReviewerID, // *int now
		ReviewNotes: ptrToString(submission.ReviewNotes),
	}, nil
}

func (s *KYCService) UpdateStatus(ctx context.Context, merchantID int, status string, reviewerID *int, notes *string) error {
	status = strings.ToLower(status)
	if status != "approved" && status != "rejected" && status != "pending" {
		return fmt.Errorf("invalid status")
	}

	latest, err := s.kycRepo.GetLatestByMerchant(ctx, merchantID) // merchantID is int
	if err != nil || latest == nil {
		return fmt.Errorf("no kyc submission found for merchant")
	}

	if err := s.kycRepo.UpdateStatus(ctx, latest.ID, status, reviewerID, notes); err != nil { // latest.ID is int, reviewerID is *int
		return err
	}

	// sync merchant KYC status
	_ = s.merchantRepo.UpdateKYCStatus(ctx, merchantID, models.KYCStatus(status)) // merchantID is int
	return nil
}

func (s *KYCService) ListByStatus(ctx context.Context, status string, limit int) ([]dto.KYCStatusResponse, error) {
	list, err := s.kycRepo.ListByStatus(ctx, status, limit)
	if err != nil {
		return nil, err
	}
	res := make([]dto.KYCStatusResponse, 0, len(list))
	for _, item := range list {
		res = append(res, dto.KYCStatusResponse{
			MerchantID:  item.MerchantID, // int
			Status:      item.Status,
			SubmittedAt: item.CreatedAt.Format(time.RFC3339),
		})
	}
	return res, nil
}

func timePtrToString(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}

func ptrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
