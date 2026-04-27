package model

import "time"

// Receipt is an immutable snapshot of an order captured at completion time.
// Stored in MongoDB — never mutated after creation.
type Receipt struct {
	ID           string         `bson:"_id,omitempty" json:"id,omitempty"`
	OrderID      int64          `bson:"order_id"      json:"order_id"`
	CompanyID    int64          `bson:"company_id"    json:"company_id"`
	OrderNo      string         `bson:"order_no"      json:"order_no"`
	CashierID    int64          `bson:"cashier_id"    json:"cashier_id"`
	CashierName  string         `bson:"cashier_name"  json:"cashier_name"`
	CustomerID   *int64         `bson:"customer_id"   json:"customer_id"`
	CustomerName *string        `bson:"customer_name" json:"customer_name"`
	Items        []ReceiptItem  `bson:"items"         json:"items"`
	Payments     []ReceiptPayment `bson:"payments"    json:"payments"`
	TotalAmount  float64        `bson:"total_amount"  json:"total_amount"`
	TotalDiscount float64       `bson:"total_discount" json:"total_discount"`
	TotalTax     float64        `bson:"total_tax"     json:"total_tax"`
	GrandTotal   float64        `bson:"grand_total"   json:"grand_total"`
	Notes        *string        `bson:"notes"         json:"notes"`
	CompletedAt  time.Time      `bson:"completed_at"  json:"completed_at"`
	CreatedAt    time.Time      `bson:"created_at"    json:"created_at"`
}

// ReceiptItem is a single line item on the receipt.
type ReceiptItem struct {
	ProductName string  `bson:"product_name" json:"product_name"`
	VariantName *string `bson:"variant_name" json:"variant_name"`
	SKU         string  `bson:"sku"          json:"sku"`
	Quantity    int     `bson:"quantity"     json:"quantity"`
	UnitPrice   float64 `bson:"unit_price"   json:"unit_price"`
	Discount    float64 `bson:"discount"     json:"discount"`
	Tax         float64 `bson:"tax"          json:"tax"`
	Subtotal    float64 `bson:"subtotal"     json:"subtotal"`
}

// ReceiptPayment is a single payment entry on the receipt.
type ReceiptPayment struct {
	Method      string  `bson:"method"       json:"method"`
	Amount      float64 `bson:"amount"       json:"amount"`
	ReferenceNo *string `bson:"reference_no" json:"reference_no"`
}
