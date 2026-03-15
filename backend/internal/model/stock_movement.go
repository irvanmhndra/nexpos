package model

import "time"

type StockMovementType string

const (
	StockMovementIn       StockMovementType = "IN"
	StockMovementOut      StockMovementType = "OUT"
	StockMovementAdjust   StockMovementType = "ADJUST"
	StockMovementTransfer StockMovementType = "TRANSFER"
)

type StockMovement struct {
	ID               int64             `db:"id"                 json:"id"`
	ProductVariantID int64             `db:"product_variant_id" json:"product_variant_id"`
	BranchID         int64             `db:"branch_id"          json:"branch_id"`
	Type             StockMovementType `db:"type"               json:"type"`
	Quantity         int               `db:"quantity"           json:"quantity"`
	StockBefore      int               `db:"stock_before"       json:"stock_before"`
	StockAfter       int               `db:"stock_after"        json:"stock_after"`
	UnitCost         *float64          `db:"unit_cost"          json:"unit_cost"`
	ReferenceType    *string           `db:"reference_type"     json:"reference_type"`
	ReferenceID      *int64            `db:"reference_id"       json:"reference_id"`
	Note             *string           `db:"note"               json:"note"`
	CreatedBy        *int64            `db:"created_by"         json:"created_by"`
	CreatedAt        time.Time         `db:"created_at"         json:"created_at"`
}
