package model

import (
	"time"

	"github.com/shopspring/decimal"
)

const (
	ShiftStatusOpen   = "open"
	ShiftStatusClosed = "closed"
)

type Shift struct {
	ID             int64            `db:"id"              json:"id"`
	CompanyID      int64            `db:"company_id"      json:"company_id"`
	BranchID       int64            `db:"branch_id"       json:"branch_id"`
	CashierID      int64            `db:"cashier_id"      json:"cashier_id"`
	Status         string           `db:"status"          json:"status"`
	OpeningFloat   decimal.Decimal  `db:"opening_float"   json:"opening_float"`
	ClosingFloat   decimal.Decimal  `db:"closing_float"   json:"closing_float"`
	ExpectedCash   decimal.Decimal  `db:"expected_cash"   json:"expected_cash"`
	ActualCash     *decimal.Decimal `db:"actual_cash"     json:"actual_cash"`
	CashDifference *decimal.Decimal `db:"cash_difference" json:"cash_difference"`
	Notes          *string          `db:"notes"           json:"notes"`
	OpenedAt       time.Time        `db:"opened_at"       json:"opened_at"`
	ClosedAt       *time.Time       `db:"closed_at"       json:"closed_at"`
	CreatedAt      time.Time        `db:"created_at"      json:"created_at"`
	UpdatedAt      time.Time        `db:"updated_at"      json:"updated_at"`
}
