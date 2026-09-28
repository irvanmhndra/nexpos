package dto

import (
	"time"

	"github.com/irvanmhndra/nexpos-api/pkg/httputil"
	"github.com/shopspring/decimal"
)

// ============== Requests ==============

type OpenShiftRequest struct {
	BranchID     int64           `json:"branch_id"     validate:"required"`
	OpeningFloat decimal.Decimal `json:"opening_float" validate:"min=0"`
}

type CloseShiftRequest struct {
	ActualCash decimal.Decimal `json:"actual_cash" validate:"min=0"`
	Notes      *string         `json:"notes"`
}

type ListShiftRequest struct {
	Page      int    `query:"page"`
	PerPage   int    `query:"per_page"`
	BranchID  *int64 `query:"branch_id"`
	CashierID *int64 `query:"cashier_id"`
	Status    string `query:"status"`
	DateFrom  string `query:"date_from"`
	DateTo    string `query:"date_to"`
}

// ============== Responses ==============

type ShiftResponse struct {
	ID             int64            `json:"id"`
	BranchID       int64            `json:"branch_id"`
	CashierID      int64            `json:"cashier_id"`
	Status         string           `json:"status"`
	OpeningFloat   decimal.Decimal  `json:"opening_float"`
	ClosingFloat   decimal.Decimal  `json:"closing_float"`
	ExpectedCash   decimal.Decimal  `json:"expected_cash"`
	ActualCash     *decimal.Decimal `json:"actual_cash"`
	CashDifference *decimal.Decimal `json:"cash_difference"`
	Notes          *string          `json:"notes"`
	OpenedAt       time.Time        `json:"opened_at"`
	ClosedAt       *time.Time       `json:"closed_at"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
}

type ShiftListResponse struct {
	Shifts     []*ShiftResponse     `json:"shifts"`
	Pagination *httputil.Pagination `json:"pagination"`
}
