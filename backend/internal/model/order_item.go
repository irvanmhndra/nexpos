package model

import "time"

type OrderItem struct {
	ID                int64     `db:"id" json:"id"`
	OrderID           int64     `db:"order_id" json:"order_id"`
	ProductID         *int64    `db:"product_id" json:"product_id"`
	ProductVariantID  *int64    `db:"product_variant_id" json:"product_variant_id"`
	SKU               string    `db:"sku" json:"sku"`
	ProductName       string    `db:"product_name" json:"product_name"`
	VariantName       *string   `db:"variant_name" json:"variant_name"`
	VariantAttributes JSONMap   `db:"variant_attributes" json:"variant_attributes"`
	UnitPrice         float64   `db:"unit_price" json:"unit_price"`
	UnitCost          float64   `db:"unit_cost" json:"unit_cost"`
	Quantity          int       `db:"quantity" json:"quantity"`
	DiscountAmount    float64   `db:"discount_amount" json:"discount_amount"`
	TaxAmount         float64   `db:"tax_amount" json:"tax_amount"`
	Subtotal          float64   `db:"subtotal" json:"subtotal"`
	CogsAmount        float64   `db:"cogs_amount" json:"cogs_amount"`
	CreatedAt         time.Time `db:"created_at" json:"created_at"`
}
