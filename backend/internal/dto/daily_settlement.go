package dto

import (
	"time"

	"github.com/irvanmhndra/nexpos-api/pkg/httputil"
	"github.com/shopspring/decimal"
)

// ============== Requests ==============

type CreateDailySettlementRequest struct {
	BranchID       int64   `json:"branch_id"       validate:"required"`
	SettlementDate string  `json:"settlement_date" validate:"required"` // YYYY-MM-DD
	Notes          *string `json:"notes"`
}

type UpdateSettlementItemRequest struct {
	ActualAmount decimal.Decimal `json:"actual_amount" validate:"gte=0"`
	Notes        *string         `json:"notes"`
}

type BulkUpdateSettlementItemsRequest struct {
	Items []BulkUpdateSettlementItem `json:"items" validate:"required,min=1,dive"`
}

type BulkUpdateSettlementItem struct {
	ItemID       int64           `json:"item_id"       validate:"required"`
	ActualAmount decimal.Decimal `json:"actual_amount" validate:"gte=0"`
	Notes        *string         `json:"notes"`
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
	ID             int64           `json:"id"`
	BranchID       int64           `json:"branch_id"`
	SettlementDate string          `json:"settlement_date"`
	Status         string          `json:"status"`
	TotalSales     decimal.Decimal `json:"total_sales"`
	TotalRefunds   decimal.Decimal `json:"total_refunds"`
	TotalExpenses  decimal.Decimal `json:"total_expenses"`
	TotalExpected  decimal.Decimal `json:"total_expected"`
	TotalActual    decimal.Decimal `json:"total_actual"`
	TotalVariance  decimal.Decimal `json:"total_variance"`
	Notes          *string         `json:"notes"`
	RecordedBy     *int64          `json:"recorded_by"`
	FinalizedBy    *int64          `json:"finalized_by"`
	FinalizedAt    *time.Time      `json:"finalized_at"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`

	Items []*DailySettlementItemResponse `json:"items,omitempty"`
}

type DailySettlementItemResponse struct {
	ID             int64            `json:"id"`
	PaymentMethod  string           `json:"payment_method"`
	ExpectedAmount decimal.Decimal  `json:"expected_amount"`
	ActualAmount   *decimal.Decimal `json:"actual_amount"`
	VarianceAmount decimal.Decimal  `json:"variance_amount"`
	Notes          *string          `json:"notes"`
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
	TotalSales     decimal.Decimal              `json:"total_sales"`
	TotalRefunds   decimal.Decimal              `json:"total_refunds"`
	TotalExpenses  decimal.Decimal              `json:"total_expenses"`
	TotalExpected  decimal.Decimal              `json:"total_expected"`
	ByMethod       []*SettlementMethodBreakdown `json:"by_method"`
}

type SettlementMethodBreakdown struct {
	PaymentMethod  string          `json:"payment_method"`
	GrossSales     decimal.Decimal `json:"gross_sales"`
	Refunds        decimal.Decimal `json:"refunds"`
	ExpensesOut    decimal.Decimal `json:"expenses_out"` // non-zero only for cash
	ExpectedAmount decimal.Decimal `json:"expected_amount"`
}
