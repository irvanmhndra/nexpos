package model

import "time"

const (
	DailySettlementStatusDraft     = "draft"
	DailySettlementStatusFinalized = "finalized"
)

type DailySettlement struct {
	ID             int64      `db:"id"              json:"id"`
	CompanyID      int64      `db:"company_id"      json:"company_id"`
	BranchID       int64      `db:"branch_id"       json:"branch_id"`
	SettlementDate string     `db:"settlement_date" json:"settlement_date"`
	Status         string     `db:"status"          json:"status"`
	TotalSales     float64    `db:"total_sales"     json:"total_sales"`
	TotalRefunds   float64    `db:"total_refunds"   json:"total_refunds"`
	TotalExpenses  float64    `db:"total_expenses"  json:"total_expenses"`
	Notes          *string    `db:"notes"           json:"notes"`
	RecordedBy     *int64     `db:"recorded_by"     json:"recorded_by"`
	FinalizedBy    *int64     `db:"finalized_by"    json:"finalized_by"`
	FinalizedAt    *time.Time `db:"finalized_at"    json:"finalized_at"`
	CreatedAt      time.Time  `db:"created_at"      json:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at"      json:"updated_at"`

	Items []*DailySettlementItem `db:"-" json:"items,omitempty"`
}

type DailySettlementItem struct {
	ID                int64     `db:"id"                  json:"id"`
	DailySettlementID int64     `db:"daily_settlement_id" json:"daily_settlement_id"`
	PaymentMethod     string    `db:"payment_method"      json:"payment_method"`
	ExpectedAmount    float64   `db:"expected_amount"     json:"expected_amount"`
	ActualAmount      *float64  `db:"actual_amount"       json:"actual_amount"`
	VarianceAmount    float64   `db:"variance_amount"     json:"variance_amount"`
	Notes             *string   `db:"notes"               json:"notes"`
	CreatedAt         time.Time `db:"created_at"          json:"created_at"`
	UpdatedAt         time.Time `db:"updated_at"          json:"updated_at"`
}
