package model

import "time"

type CompanySettings struct {
	ID        int64 `db:"id" json:"id"`
	CompanyID int64 `db:"company_id" json:"company_id"`

	// Tax configuration
	TaxEnabled   bool    `db:"tax_enabled" json:"tax_enabled"`
	TaxRate      float64 `db:"tax_rate" json:"tax_rate"`
	TaxInclusive bool    `db:"tax_inclusive" json:"tax_inclusive"`

	// Rounding configuration
	RoundingEnabled bool    `db:"rounding_enabled" json:"rounding_enabled"`
	RoundingAmount  float64 `db:"rounding_amount" json:"rounding_amount"`

	// Order settings
	AutoCompleteCounterOrders  bool `db:"auto_complete_counter_orders" json:"auto_complete_counter_orders"`
	RequireCustomerForDelivery bool `db:"require_customer_for_delivery" json:"require_customer_for_delivery"`

	// Receipt settings
	ReceiptHeader    *string `db:"receipt_header" json:"receipt_header"`
	ReceiptFooter    *string `db:"receipt_footer" json:"receipt_footer"`
	ShowTaxOnReceipt bool    `db:"show_tax_on_receipt" json:"show_tax_on_receipt"`

	// Offline settings
	OfflineModeEnabled bool `db:"offline_mode_enabled" json:"offline_mode_enabled"`
	MaxOfflineDays     int  `db:"max_offline_days" json:"max_offline_days"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
