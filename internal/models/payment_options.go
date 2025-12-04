package models

import (
	"time"
)

type FeeType string

const (
	FeeTypePercentage FeeType = "percentage"
	FeeTypeFlat       FeeType = "flat"
)

type PaymentOptions struct {
	ID                      string    `json:"id" db:"id"`
	MerchantID              string    `json:"merchant_id" db:"merchant_id"`
	CardEnabled             bool      `json:"card_enabled" db:"card_enabled"`
	BankTransferEnabled     bool      `json:"bank_transfer_enabled" db:"bank_transfer_enabled"`
	USSDEnabled             bool      `json:"ussd_enabled" db:"ussd_enabled"`
	QREnabled               bool      `json:"qr_enabled" db:"qr_enabled"`
	MobileMoneyEnabled      bool      `json:"mobile_money_enabled" db:"mobile_money_enabled"`
	CardFeeType             FeeType   `json:"card_fee_type" db:"card_fee_type"`
	CardFeeValue            float64   `json:"card_fee_value" db:"card_fee_value"`
	CardCap                 *int64    `json:"card_cap,omitempty" db:"card_cap"`
	BankTransferFeeType     FeeType   `json:"bank_transfer_fee_type" db:"bank_transfer_fee_type"`
	BankTransferFeeValue    float64   `json:"bank_transfer_fee_value" db:"bank_transfer_fee_value"`
	USSDFeeType             FeeType   `json:"ussd_fee_type" db:"ussd_fee_type"`
	USSDFeeValue            float64   `json:"ussd_fee_value" db:"ussd_fee_value"`
	CreatedAt               time.Time `json:"created_at" db:"created_at"`
	UpdatedAt               time.Time `json:"updated_at" db:"updated_at"`
}

// CalculateFee calculates the fee for a given amount based on payment method
func (po *PaymentOptions) CalculateFee(amount int64, paymentMethod string) int64 {
	var feeType FeeType
	var feeValue float64
	var cap *int64

	switch paymentMethod {
	case "card":
		if !po.CardEnabled {
			return 0
		}
		feeType = po.CardFeeType
		feeValue = po.CardFeeValue
		cap = po.CardCap
	case "bank_transfer":
		if !po.BankTransferEnabled {
			return 0
		}
		feeType = po.BankTransferFeeType
		feeValue = po.BankTransferFeeValue
	case "ussd":
		if !po.USSDEnabled {
			return 0
		}
		feeType = po.USSDFeeType
		feeValue = po.USSDFeeValue
	default:
		return 0
	}

	var fee int64
	if feeType == FeeTypePercentage {
		// Calculate percentage fee
		fee = int64(float64(amount) * feeValue / 100.0)
	} else {
		// Flat fee (feeValue is in naira, convert to kobo)
		fee = int64(feeValue * 100)
	}

	// Apply cap if exists
	if cap != nil && fee > *cap {
		fee = *cap
	}

	return fee
}

// GetEnabledMethods returns a list of enabled payment methods
func (po *PaymentOptions) GetEnabledMethods() []string {
	methods := []string{}
	if po.CardEnabled {
		methods = append(methods, "card")
	}
	if po.BankTransferEnabled {
		methods = append(methods, "bank_transfer")
	}
	if po.USSDEnabled {
		methods = append(methods, "ussd")
	}
	if po.QREnabled {
		methods = append(methods, "qr")
	}
	if po.MobileMoneyEnabled {
		methods = append(methods, "mobile_money")
	}
	return methods
}
