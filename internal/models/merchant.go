package models

import "time"

// KYCStatus represents the verification status of a merchant's KYC
type KYCStatus string

const (
	KYCStatusPending  KYCStatus = "pending"
	KYCStatusApproved KYCStatus = "approved"
	KYCStatusRejected KYCStatus = "rejected"
	KYCStatusNotStarted KYCStatus = "not_started"
)

// MerchantStatus represents the overall status of a merchant account
type MerchantStatus string

const (
	MerchantStatusActive    MerchantStatus = "active"
	MerchantStatusSuspended MerchantStatus = "suspended"
	MerchantStatusInactive  MerchantStatus = "inactive"
)

// Merchant represents a merchant entity in the system
type Merchant struct {
	ID           int            `json:"id"`
	Name         string         `json:"name"`
	Email        string         `json:"email"`
	BusinessName string         `json:"business_name"`
	Country      string         `json:"country"`
	Status       MerchantStatus `json:"status"`
	KYCStatus    KYCStatus      `json:"kyc_status"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

// CanTransact checks if a merchant is allowed to process transactions
func (m *Merchant) CanTransact() bool {
	// Merchant must be active and have approved KYC to transact
	return m.Status == MerchantStatusActive && m.KYCStatus == KYCStatusApproved
}

// IsKYCCompleted checks if KYC process is completed (approved or rejected)
func (m *Merchant) IsKYCCompleted() bool {
	return m.KYCStatus == KYCStatusApproved || m.KYCStatus == KYCStatusRejected
}
