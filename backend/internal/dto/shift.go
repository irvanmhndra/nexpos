package dto

import (
	"time"

	"github.com/irvanmhndra/nexpos-api/pkg/httputil"
)

// ============== Requests ==============

type OpenShiftRequest struct {
	BranchID     int64   `json:"branch_id"     validate:"required"`
	OpeningFloat float64 `json:"opening_float" validate:"min=0"`
}

type CloseShiftRequest struct {
	ActualCash float64 `json:"actual_cash" validate:"min=0"`
	Notes      *string `json:"notes"`
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
	ID             int64      `json:"id"`
	BranchID       int64      `json:"branch_id"`
	CashierID      int64      `json:"cashier_id"`
	Status         string     `json:"status"`
	OpeningFloat   float64    `json:"opening_float"`
	ClosingFloat   float64    `json:"closing_float"`
	ExpectedCash   float64    `json:"expected_cash"`
	ActualCash     *float64   `json:"actual_cash"`
	CashDifference *float64   `json:"cash_difference"`
	Notes          *string    `json:"notes"`
	OpenedAt       time.Time  `json:"opened_at"`
	ClosedAt       *time.Time `json:"closed_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type ShiftListResponse struct {
	Shifts     []*ShiftResponse     `json:"shifts"`
	Pagination *httputil.Pagination `json:"pagination"`
}
