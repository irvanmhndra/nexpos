package repository

import (
	"time"

	"github.com/irvanmhndra/nexpos-api/internal/model"
)

// InventoryRow is the result type for inventory list queries
type InventoryRow struct {
	ProductVariantID int64   `db:"product_variant_id"`
	ProductID        int64   `db:"product_id"`
	ProductName      string  `db:"product_name"`
	VariantName      string  `db:"variant_name"`
	SKU              string  `db:"sku"`
	CategoryName     string  `db:"category_name"`
	ImageData        *string `db:"image_data"`
	BranchID         int64   `db:"branch_id"`
	BranchName       string  `db:"branch_name"`
	CurrentStock     int     `db:"current_stock"`
	MinStock         int     `db:"min_stock"`
	StockValue       float64 `db:"stock_value"`
	HasVariants      bool    `db:"has_variants"`
}

// InventoryStats is the result type for inventory stats queries
type InventoryStats struct {
	TotalSKU        int     `db:"total_sku"`
	LowStock        int     `db:"low_stock"`
	OutOfStock      int     `db:"out_of_stock"`
	TotalStockValue float64 `db:"total_stock_value"`
}

// MovementRow is the result type for stock movement list queries
type MovementRow struct {
	ID               int64                   `db:"id"`
	ProductVariantID int64                   `db:"product_variant_id"`
	ProductName      string                  `db:"product_name"`
	VariantName      string                  `db:"variant_name"`
	SKU              string                  `db:"sku"`
	BranchID         int64                   `db:"branch_id"`
	BranchName       string                  `db:"branch_name"`
	Type             model.StockMovementType `db:"type"`
	Quantity         int                     `db:"quantity"`
	StockBefore      int                     `db:"stock_before"`
	StockAfter       int                     `db:"stock_after"`
	UnitCost         *float64                `db:"unit_cost"`
	ReferenceType    *string                 `db:"reference_type"`
	ReferenceID      *int64                  `db:"reference_id"`
	Note             *string                 `db:"note"`
	CreatedBy        *int64                  `db:"created_by"`
	CreatedByName    string                  `db:"created_by_name"`
	CreatedAt        time.Time               `db:"created_at"`
}

// MovementStats is the result type for movement stats queries
type MovementStats struct {
	TotalMovements int `db:"total_movements"`
	StockIn        int `db:"stock_in"`
	StockOut       int `db:"stock_out"`
	Adjustments    int `db:"adjustments"`
}
