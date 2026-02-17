package dto

import "time"

type CompanySettingsResponse struct {
	ID        int64 `json:"id"`
	CompanyID int64 `json:"company_id"`

	// Tax configuration
	TaxEnabled   bool    `json:"tax_enabled"`
	TaxRate      float64 `json:"tax_rate"`
	TaxInclusive bool    `json:"tax_inclusive"`

	// Rounding configuration
	RoundingEnabled bool    `json:"rounding_enabled"`
	RoundingAmount  float64 `json:"rounding_amount"`

	// Order settings
	AutoCompleteCounterOrders  bool `json:"auto_complete_counter_orders"`
	RequireCustomerForDelivery bool `json:"require_customer_for_delivery"`

	// Receipt settings
	ReceiptHeader    *string `json:"receipt_header"`
	ReceiptFooter    *string `json:"receipt_footer"`
	ShowTaxOnReceipt bool    `json:"show_tax_on_receipt"`

	// Offline settings
	OfflineModeEnabled bool `json:"offline_mode_enabled"`
	MaxOfflineDays     int  `json:"max_offline_days"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UpdateCompanySettingsRequest struct {
	// Tax configuration
	TaxEnabled   *bool    `json:"tax_enabled"`
	TaxRate      *float64 `json:"tax_rate" validate:"omitempty,gte=0,lte=100"`
	TaxInclusive *bool    `json:"tax_inclusive"`

	// Rounding configuration
	RoundingEnabled *bool    `json:"rounding_enabled"`
	RoundingAmount  *float64 `json:"rounding_amount" validate:"omitempty,gte=0"`

	// Order settings
	AutoCompleteCounterOrders  *bool `json:"auto_complete_counter_orders"`
	RequireCustomerForDelivery *bool `json:"require_customer_for_delivery"`

	// Receipt settings
	ReceiptHeader    *string `json:"receipt_header"`
	ReceiptFooter    *string `json:"receipt_footer"`
	ShowTaxOnReceipt *bool   `json:"show_tax_on_receipt"`

	// Offline settings
	OfflineModeEnabled *bool `json:"offline_mode_enabled"`
	MaxOfflineDays     *int  `json:"max_offline_days" validate:"omitempty,gte=1,lte=30"`
}
