package dto

import (
	"time"

	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/irvanmhndra/nexpos-api/pkg/httputil"
)

// ============== Stock Status ==============

type StockStatus string

const (
	StockStatusNormal     StockStatus = "normal"
	StockStatusLow        StockStatus = "low"
	StockStatusOutOfStock StockStatus = "out_of_stock"
)

// ============== Requests ==============

type AdjustStockRequest struct {
	VariantID int64                    `json:"variant_id" validate:"required"`
	BranchID  int64                    `json:"branch_id"  validate:"required"`
	Type      model.StockMovementType  `json:"type"       validate:"required,oneof=IN OUT ADJUST"`
	Quantity  int                      `json:"quantity"   validate:"required,min=1"`
	UnitCost  *float64                 `json:"unit_cost"`
	Note      *string                  `json:"note"`
}

type UpdateMinStockRequest struct {
	MinQuantity int `json:"min_quantity" validate:"min=0"`
}

type ListInventoryRequest struct {
	Page     int    `query:"page"`
	PerPage  int    `query:"per_page"`
	Search   string `query:"search"`
	Category string `query:"category"`
	Status   string `query:"status"`
	BranchID int64  `query:"branch_id"`
}

type ListMovementsRequest struct {
	Page      int    `query:"page"`
	PerPage   int    `query:"per_page"`
	Search    string `query:"search"`
	Type      string `query:"type"`
	StartDate string `query:"start_date"`
	EndDate   string `query:"end_date"`
	BranchID  int64  `query:"branch_id"`
}

// ============== Responses ==============

type InventoryItemResponse struct {
	ProductVariantID int64       `json:"product_variant_id"`
	ProductID        int64       `json:"product_id"`
	ProductName      string      `json:"product_name"`
	VariantName      string      `json:"variant_name"`
	SKU              string      `json:"sku"`
	CategoryName     string      `json:"category_name"`
	ImageData        *string     `json:"image_data"`
	BranchID         int64       `json:"branch_id"`
	BranchName       string      `json:"branch_name"`
	CurrentStock     int         `json:"current_stock"`
	MinStock         int         `json:"min_stock"`
	StockValue       float64     `json:"stock_value"`
	Status           StockStatus `json:"status"`
	HasVariants      bool        `json:"has_variants"`
}

type InventoryStatsResponse struct {
	TotalSKU        int     `json:"total_sku"`
	LowStock        int     `json:"low_stock"`
	OutOfStock      int     `json:"out_of_stock"`
	TotalStockValue float64 `json:"total_stock_value"`
}

type StockMovementResponse struct {
	ID               int64                   `json:"id"`
	ProductVariantID int64                   `json:"product_variant_id"`
	ProductName      string                  `json:"product_name"`
	VariantName      string                  `json:"variant_name"`
	SKU              string                  `json:"sku"`
	BranchID         int64                   `json:"branch_id"`
	BranchName       string                  `json:"branch_name"`
	Type             model.StockMovementType `json:"type"`
	Quantity         int                     `json:"quantity"`
	StockBefore      int                     `json:"stock_before"`
	StockAfter       int                     `json:"stock_after"`
	UnitCost         *float64                `json:"unit_cost"`
	ReferenceType    *string                 `json:"reference_type"`
	ReferenceID      *int64                  `json:"reference_id"`
	Note             *string                 `json:"note"`
	CreatedBy        *int64                  `json:"created_by"`
	CreatedByName    string                  `json:"created_by_name"`
	CreatedAt        time.Time               `json:"created_at"`
}

type MovementStatsResponse struct {
	TotalMovements int `json:"total_movements"`
	StockIn        int `json:"stock_in"`
	StockOut       int `json:"stock_out"`
	Adjustments    int `json:"adjustments"`
}

type InventoryListResponse struct {
	Items      []*InventoryItemResponse `json:"items"`
	Pagination *httputil.Pagination     `json:"pagination"`
}

type MovementListResponse struct {
	Movements  []*StockMovementResponse `json:"movements"`
	Pagination *httputil.Pagination     `json:"pagination"`
}
