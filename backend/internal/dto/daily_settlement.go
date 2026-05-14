package dto

import (
	"time"

	"github.com/irvanmhndra/nexpos-api/pkg/httputil"
)

// ============== Requests ==============

type CreateDailySettlementRequest struct {
	BranchID       int64   `json:"branch_id"       validate:"required"`
	SettlementDate string  `json:"settlement_date" validate:"required"` // YYYY-MM-DD
	Notes          *string `json:"notes"`
}

type UpdateSettlementItemRequest struct {
	ActualAmount float64 `json:"actual_amount" validate:"gte=0"`
	Notes        *string `json:"notes"`
}

type BulkUpdateSettlementItemsRequest struct {
	Items []BulkUpdateSettlementItem `json:"items" validate:"required,min=1,dive"`
}

type BulkUpdateSettlementItem struct {
	ItemID       int64   `json:"item_id"       validate:"required"`
	ActualAmount float64 `json:"actual_amount" validate:"gte=0"`
	Notes        *string `json:"notes"`
}

type FinalizeDailySettlementRequest struct {
	Notes *string `json:"notes"`
}

type ListDailySettlementRequest struct {
	Page     int    `query:"page"`
	PerPage  int    `query:"per_page"`
	Status   string `query:"status"`
	BranchID *int64 `query:"branch_id"`
	DateFrom string `query:"date_from"`
	DateTo   string `query:"date_to"`
}

type DailySettlementReportRequest struct {
	BranchID int64  `query:"branch_id"`
	Date     string `query:"date"`
}

// ============== Responses ==============

type DailySettlementResponse struct {
	ID             int64      `json:"id"`
	BranchID       int64      `json:"branch_id"`
	SettlementDate string     `json:"settlement_date"`
	Status         string     `json:"status"`
	TotalSales     float64    `json:"total_sales"`
	TotalRefunds   float64    `json:"total_refunds"`
	TotalExpenses  float64    `json:"total_expenses"`
	TotalExpected  float64    `json:"total_expected"`
	TotalActual    float64    `json:"total_actual"`
	TotalVariance  float64    `json:"total_variance"`
	Notes          *string    `json:"notes"`
	RecordedBy     *int64     `json:"recorded_by"`
	FinalizedBy    *int64     `json:"finalized_by"`
	FinalizedAt    *time.Time `json:"finalized_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`

	Items []*DailySettlementItemResponse `json:"items,omitempty"`
}

type DailySettlementItemResponse struct {
	ID             int64    `json:"id"`
	PaymentMethod  string   `json:"payment_method"`
	ExpectedAmount float64  `json:"expected_amount"`
	ActualAmount   *float64 `json:"actual_amount"`
	VarianceAmount float64  `json:"variance_amount"`
	Notes          *string  `json:"notes"`
}

type DailySettlementListResponse struct {
	Settlements []*DailySettlementResponse `json:"settlements"`
	Pagination  *httputil.Pagination       `json:"pagination"`
}

// DailySettlementReportResponse is a preview (no persistence) of expected
// amounts for a given branch + date. Useful for browsing days before
// recording an actual settlement.
type DailySettlementReportResponse struct {
	BranchID       int64                        `json:"branch_id"`
	SettlementDate string                       `json:"settlement_date"`
	TotalSales     float64                      `json:"total_sales"`
	TotalRefunds   float64                      `json:"total_refunds"`
	TotalExpenses  float64                      `json:"total_expenses"`
	TotalExpected  float64                      `json:"total_expected"`
	ByMethod       []*SettlementMethodBreakdown `json:"by_method"`
}

type SettlementMethodBreakdown struct {
	PaymentMethod  string  `json:"payment_method"`
	GrossSales     float64 `json:"gross_sales"`
	Refunds        float64 `json:"refunds"`
	ExpensesOut    float64 `json:"expenses_out"` // non-zero only for cash
	ExpectedAmount float64 `json:"expected_amount"`
}
