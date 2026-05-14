package dto

import (
	"time"

	"github.com/irvanmhndra/nexpos-api/pkg/httputil"
)

// ============== Requests ==============

type CreateStockOpnameRequest struct {
	BranchID   int64   `json:"branch_id"   validate:"required"`
	Notes      *string `json:"notes"`
	CategoryID *int64  `json:"category_id"`
}

type UpdateOpnameItemRequest struct {
	CountedStock int     `json:"counted_stock" validate:"min=0"`
	Notes        *string `json:"notes"`
}

type BulkUpdateOpnameItemsRequest struct {
	Items []BulkUpdateOpnameItem `json:"items" validate:"required,min=1,dive"`
}

type BulkUpdateOpnameItem struct {
	ItemID       int64   `json:"item_id"       validate:"required"`
	CountedStock int     `json:"counted_stock" validate:"min=0"`
	Notes        *string `json:"notes"`
}

type CompleteStockOpnameRequest struct {
	Notes *string `json:"notes"`
}

type ListStockOpnameRequest struct {
	Page     int    `query:"page"`
	PerPage  int    `query:"per_page"`
	Status   string `query:"status"`
	BranchID *int64 `query:"branch_id"`
}

// ============== Responses ==============

type StockOpnameResponse struct {
	ID                 int64      `json:"id"`
	BranchID           int64      `json:"branch_id"`
	OpnameNumber       string     `json:"opname_number"`
	Status             string     `json:"status"`
	Notes              *string    `json:"notes"`
	TotalItems         int        `json:"total_items"`
	CountedItems       int        `json:"counted_items"`
	TotalVarianceQty   int        `json:"total_variance_qty"`
	TotalVarianceValue float64    `json:"total_variance_value"`
	StartedBy          *int64     `json:"started_by"`
	CompletedBy        *int64     `json:"completed_by"`
	StartedAt          time.Time  `json:"started_at"`
	CompletedAt        *time.Time `json:"completed_at"`
	CancelledAt        *time.Time `json:"cancelled_at"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`

	Items []*StockOpnameItemResponse `json:"items,omitempty"`
}

type StockOpnameItemResponse struct {
	ID               int64      `json:"id"`
	StockOpnameID    int64      `json:"stock_opname_id"`
	ProductVariantID int64      `json:"product_variant_id"`
	SKU              string     `json:"sku"`
	ProductName      string     `json:"product_name"`
	VariantName      string     `json:"variant_name"`
	SystemStock      int        `json:"system_stock"`
	CountedStock     *int       `json:"counted_stock"`
	VarianceQty      int        `json:"variance_qty"`
	UnitCost         float64    `json:"unit_cost"`
	VarianceValue    float64    `json:"variance_value"`
	Notes            *string    `json:"notes"`
	CountedAt        *time.Time `json:"counted_at"`
}

type StockOpnameListResponse struct {
	Opnames    []*StockOpnameResponse `json:"opnames"`
	Pagination *httputil.Pagination   `json:"pagination"`
}
