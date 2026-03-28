package dto

import (
	"time"

	"github.com/irvanmhndra/nexpos-api/pkg/httputil"
)

// ============== Requests ==============

type CreatePOItemRequest struct {
	ProductVariantID *int64  `json:"product_variant_id"`
	SKU              string  `json:"sku"          validate:"required,max=100"`
	VariantName      string  `json:"variant_name" validate:"required,max=255"`
	Quantity         int     `json:"quantity"     validate:"required,min=1"`
	UnitCost         float64 `json:"unit_cost"    validate:"min=0"`
}

type CreatePurchaseOrderRequest struct {
	BranchID   int64                 `json:"branch_id"   validate:"required"`
	SupplierID int64                 `json:"supplier_id" validate:"required"`
	Notes      *string               `json:"notes"`
	Items      []CreatePOItemRequest `json:"items"       validate:"required,min=1,dive"`
}

type ReceivePOItemRequest struct {
	ItemID           int64 `json:"item_id"           validate:"required"`
	ReceivedQuantity int   `json:"received_quantity" validate:"min=0"`
}

type ReceivePurchaseOrderRequest struct {
	Items []ReceivePOItemRequest `json:"items" validate:"required,min=1,dive"`
}

type ListPurchaseOrderRequest struct {
	Page       int    `query:"page"`
	PerPage    int    `query:"per_page"`
	Status     string `query:"status"`
	SupplierID *int64 `query:"supplier_id"`
}

// ============== Responses ==============

type POItemResponse struct {
	ID               int64   `json:"id"`
	ProductVariantID *int64  `json:"product_variant_id"`
	SKU              string  `json:"sku"`
	VariantName      string  `json:"variant_name"`
	Quantity         int     `json:"quantity"`
	ReceivedQuantity int     `json:"received_quantity"`
	UnitCost         float64 `json:"unit_cost"`
	Subtotal         float64 `json:"subtotal"`
}

type PurchaseOrderResponse struct {
	ID          int64           `json:"id"`
	BranchID    int64           `json:"branch_id"`
	SupplierID  int64           `json:"supplier_id"`
	PONumber    string          `json:"po_number"`
	Status      string          `json:"status"`
	Notes       *string         `json:"notes"`
	TotalAmount float64         `json:"total_amount"`
	OrderedAt   *time.Time      `json:"ordered_at"`
	ReceivedAt  *time.Time      `json:"received_at"`
	CancelledAt *time.Time      `json:"cancelled_at"`
	Items       []*POItemResponse `json:"items,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type PurchaseOrderListResponse struct {
	PurchaseOrders []*PurchaseOrderResponse `json:"purchase_orders"`
	Pagination     *httputil.Pagination     `json:"pagination"`
}
