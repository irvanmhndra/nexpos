package model

import "time"

const (
	POStatusDraft     = "draft"
	POStatusOrdered   = "ordered"
	POStatusPartial   = "partial"
	POStatusReceived  = "received"
	POStatusCancelled = "cancelled"
)

type PurchaseOrder struct {
	ID          int64      `db:"id"           json:"id"`
	CompanyID   int64      `db:"company_id"   json:"company_id"`
	BranchID    int64      `db:"branch_id"    json:"branch_id"`
	SupplierID  int64      `db:"supplier_id"  json:"supplier_id"`
	PONumber    string     `db:"po_number"    json:"po_number"`
	Status      string     `db:"status"       json:"status"`
	Notes       *string    `db:"notes"        json:"notes"`
	TotalAmount float64    `db:"total_amount" json:"total_amount"`
	OrderedAt   *time.Time `db:"ordered_at"   json:"ordered_at"`
	ReceivedAt  *time.Time `db:"received_at"  json:"received_at"`
	CancelledAt *time.Time `db:"cancelled_at" json:"cancelled_at"`
	CreatedAt   time.Time  `db:"created_at"   json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"   json:"updated_at"`

	Items []*PurchaseOrderItem `db:"-" json:"items,omitempty"`
}

type PurchaseOrderItem struct {
	ID               int64   `db:"id"                  json:"id"`
	PurchaseOrderID  int64   `db:"purchase_order_id"   json:"purchase_order_id"`
	ProductVariantID *int64  `db:"product_variant_id"  json:"product_variant_id"`
	SKU              string  `db:"sku"                 json:"sku"`
	VariantName      string  `db:"variant_name"        json:"variant_name"`
	Quantity         int     `db:"quantity"            json:"quantity"`
	ReceivedQuantity int     `db:"received_quantity"   json:"received_quantity"`
	UnitCost         float64 `db:"unit_cost"           json:"unit_cost"`
	Subtotal         float64 `db:"subtotal"            json:"subtotal"`
	CreatedAt        time.Time `db:"created_at"        json:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"        json:"updated_at"`
}
