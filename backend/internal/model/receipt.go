package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// Receipt is an immutable snapshot of an order captured at completion time.
// Stored as a single Postgres row with items and payments as JSONB.
type Receipt struct {
	ID            int64           `db:"id"             json:"id"`
	CompanyID     int64           `db:"company_id"     json:"company_id"`
	OrderID       int64           `db:"order_id"       json:"order_id"`
	OrderNo       string          `db:"order_no"       json:"order_no"`
	CashierID     int64           `db:"cashier_id"     json:"cashier_id"`
	CashierName   string          `db:"cashier_name"   json:"cashier_name"`
	CustomerID    *int64          `db:"customer_id"    json:"customer_id"`
	CustomerName  *string         `db:"customer_name"  json:"customer_name"`
	Items         ReceiptItems    `db:"items"          json:"items"`
	Payments      ReceiptPayments `db:"payments"       json:"payments"`
	TotalAmount   float64         `db:"total_amount"   json:"total_amount"`
	TotalDiscount float64         `db:"total_discount" json:"total_discount"`
	TotalTax      float64         `db:"total_tax"      json:"total_tax"`
	GrandTotal    float64         `db:"grand_total"    json:"grand_total"`
	Notes         *string         `db:"notes"          json:"notes"`
	CompletedAt   time.Time       `db:"completed_at"   json:"completed_at"`
	CreatedAt     time.Time       `db:"created_at"     json:"created_at"`
}

// ReceiptItem is a single line item on the receipt.
type ReceiptItem struct {
	ProductName string  `json:"product_name"`
	VariantName *string `json:"variant_name"`
	SKU         string  `json:"sku"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	Discount    float64 `json:"discount"`
	Tax         float64 `json:"tax"`
	Subtotal    float64 `json:"subtotal"`
}

// ReceiptPayment is a single payment entry on the receipt.
type ReceiptPayment struct {
	Method      string  `json:"method"`
	Amount      float64 `json:"amount"`
	ReferenceNo *string `json:"reference_no"`
}

// ReceiptItems is a JSONB-backed slice. Implements driver.Valuer + sql.Scanner.
type ReceiptItems []ReceiptItem

func (r ReceiptItems) Value() (driver.Value, error) {
	if r == nil {
		return []byte("[]"), nil
	}
	return json.Marshal(r)
}

func (r *ReceiptItems) Scan(value interface{}) error {
	if value == nil {
		*r = nil
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return errors.New("ReceiptItems.Scan: expected []byte")
	}
	return json.Unmarshal(b, r)
}

// ReceiptPayments is a JSONB-backed slice.
type ReceiptPayments []ReceiptPayment

func (r ReceiptPayments) Value() (driver.Value, error) {
	if r == nil {
		return []byte("[]"), nil
	}
	return json.Marshal(r)
}

func (r *ReceiptPayments) Scan(value interface{}) error {
	if value == nil {
		*r = nil
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return errors.New("ReceiptPayments.Scan: expected []byte")
	}
	return json.Unmarshal(b, r)
}
