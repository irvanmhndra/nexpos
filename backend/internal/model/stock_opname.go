package model

import "time"

const (
	StockOpnameStatusInProgress = "in_progress"
	StockOpnameStatusCompleted  = "completed"
	StockOpnameStatusCancelled  = "cancelled"
)

type StockOpname struct {
	ID                 int64      `db:"id"                   json:"id"`
	CompanyID          int64      `db:"company_id"           json:"company_id"`
	BranchID           int64      `db:"branch_id"            json:"branch_id"`
	OpnameNumber       string     `db:"opname_number"        json:"opname_number"`
	Status             string     `db:"status"               json:"status"`
	Notes              *string    `db:"notes"                json:"notes"`
	TotalVarianceQty   int        `db:"total_variance_qty"   json:"total_variance_qty"`
	TotalVarianceValue float64    `db:"total_variance_value" json:"total_variance_value"`
	StartedBy          *int64     `db:"started_by"           json:"started_by"`
	CompletedBy        *int64     `db:"completed_by"         json:"completed_by"`
	StartedAt          time.Time  `db:"started_at"           json:"started_at"`
	CompletedAt        *time.Time `db:"completed_at"         json:"completed_at"`
	CancelledAt        *time.Time `db:"cancelled_at"         json:"cancelled_at"`
	CreatedAt          time.Time  `db:"created_at"           json:"created_at"`
	UpdatedAt          time.Time  `db:"updated_at"           json:"updated_at"`

	Items []*StockOpnameItem `db:"-" json:"items,omitempty"`
}

type StockOpnameItem struct {
	ID               int64      `db:"id"                 json:"id"`
	StockOpnameID    int64      `db:"stock_opname_id"    json:"stock_opname_id"`
	ProductVariantID int64      `db:"product_variant_id" json:"product_variant_id"`
	SKU              string     `db:"sku"                json:"sku"`
	ProductName      string     `db:"product_name"       json:"product_name"`
	VariantName      string     `db:"variant_name"       json:"variant_name"`
	SystemStock      int        `db:"system_stock"       json:"system_stock"`
	CountedStock     *int       `db:"counted_stock"      json:"counted_stock"`
	VarianceQty      int        `db:"variance_qty"       json:"variance_qty"`
	UnitCost         float64    `db:"unit_cost"          json:"unit_cost"`
	VarianceValue    float64    `db:"variance_value"     json:"variance_value"`
	Notes            *string    `db:"notes"              json:"notes"`
	CountedBy        *int64     `db:"counted_by"         json:"counted_by"`
	CountedAt        *time.Time `db:"counted_at"         json:"counted_at"`
	CreatedAt        time.Time  `db:"created_at"         json:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at"         json:"updated_at"`
}
