package models

import (
	"time"
)

type ScheduleType string

const (
	ScheduleTypeDaily  ScheduleType = "daily"
	ScheduleTypeWeekly ScheduleType = "weekly"
	ScheduleTypeManual ScheduleType = "manual"
)

type SettlementConfig struct {
	ID                  string       `json:"id" db:"id"`
	MerchantID          string       `json:"merchant_id" db:"merchant_id"`
	ScheduleType        ScheduleType `json:"schedule_type" db:"schedule_type"`
	SettlementTime      string       `json:"settlement_time" db:"settlement_time"` // HH:MM:SS format
	SettlementDays      []int        `json:"settlement_days" db:"settlement_days"` // Days of week (1=Mon, 7=Sun)
	MinimumAmount       int64        `json:"minimum_amount" db:"minimum_amount"`   // Minimum balance in kobo
	AutoSettle          bool         `json:"auto_settle" db:"auto_settle"`
	SettlementDelayDays int          `json:"settlement_delay_days" db:"settlement_delay_days"` // T+N days
	Currency            string       `json:"currency" db:"currency"`
	CreatedAt           time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time    `json:"updated_at" db:"updated_at"`
}

// ShouldSettleToday determines if settlement should run today
func (sc *SettlementConfig) ShouldSettleToday() bool {
	if !sc.AutoSettle {
		return false
	}

	if sc.ScheduleType == ScheduleTypeManual {
		return false
	}

	today := time.Now().Weekday()
	dayNum := int(today)
	if dayNum == 0 { // Sunday
		dayNum = 7
	}

	if sc.ScheduleType == ScheduleTypeDaily {
		return true
	}

	// Weekly - check if today is in settlement days
	for _, day := range sc.SettlementDays {
		if day == dayNum {
			return true
		}
	}

	return false
}

// GetNextSettlementTime returns the next settlement time
func (sc *SettlementConfig) GetNextSettlementTime() time.Time {
	now := time.Now()

	// Parse settlement time
	settlementTime, err := time.Parse("15:04:05", sc.SettlementTime)
	if err != nil {
		// Default to 9 AM if parsing fails
		settlementTime = time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC)
	}

	// Set settlement time for today
	nextSettlement := time.Date(
		now.Year(), now.Month(), now.Day(),
		settlementTime.Hour(), settlementTime.Minute(), settlementTime.Second(),
		0, now.Location(),
	)

	// If time has passed today, move to next valid day
	if nextSettlement.Before(now) {
		nextSettlement = nextSettlement.Add(24 * time.Hour)
	}

	// For weekly schedules, find next valid day
	if sc.ScheduleType == ScheduleTypeWeekly {
		for i := 0; i < 7; i++ {
			dayNum := int(nextSettlement.Weekday())
			if dayNum == 0 {
				dayNum = 7
			}

			for _, validDay := range sc.SettlementDays {
				if dayNum == validDay {
					return nextSettlement
				}
			}

			nextSettlement = nextSettlement.Add(24 * time.Hour)
		}
	}

	return nextSettlement
}

// CalculateSettlementDate calculates when funds from a transaction should be settled
// based on the transaction date and settlement delay
func (sc *SettlementConfig) CalculateSettlementDate(transactionDate time.Time) time.Time {
	settlementDate := transactionDate.AddDate(0, 0, sc.SettlementDelayDays)

	// If not auto-settle or manual, return the date as-is
	if !sc.AutoSettle || sc.ScheduleType == ScheduleTypeManual {
		return settlementDate
	}

	// For daily schedules, settle on the calculated date
	if sc.ScheduleType == ScheduleTypeDaily {
		return settlementDate
	}

	// For weekly schedules, find the next valid settlement day
	for i := 0; i < 7; i++ {
		dayNum := int(settlementDate.Weekday())
		if dayNum == 0 {
			dayNum = 7
		}

		for _, validDay := range sc.SettlementDays {
			if dayNum == validDay {
				return settlementDate
			}
		}

		settlementDate = settlementDate.AddDate(0, 0, 1)
	}

	return settlementDate
}
