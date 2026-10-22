package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ============== Order Item DTOs ==============

type OrderItemInput struct {
	ProductVariantID *int64          `json:"product_variant_id"`
	Quantity         int             `json:"quantity" validate:"required,min=1"`
	DiscountAmount   decimal.Decimal `json:"discount_amount" validate:"gte=0"`
	Notes            *string         `json:"notes"`
}

type OrderItemResponse struct {
	ID                int64                  `json:"id"`
	ProductID         *int64                 `json:"product_id"`
	ProductVariantID  *int64                 `json:"product_variant_id"`
	SKU               string                 `json:"sku"`
	ProductName       string                 `json:"product_name"`
	VariantName       *string                `json:"variant_name"`
	VariantAttributes map[string]interface{} `json:"variant_attributes"`
	UnitPrice         decimal.Decimal        `json:"unit_price"`
	UnitCost          decimal.Decimal        `json:"unit_cost"`
	Quantity          int                    `json:"quantity"`
	DiscountAmount    decimal.Decimal        `json:"discount_amount"`
	TaxAmount         decimal.Decimal        `json:"tax_amount"`
	Subtotal          decimal.Decimal        `json:"subtotal"`
	CogsAmount        decimal.Decimal        `json:"cogs_amount"`
}

// ============== Payment DTOs ==============

type PaymentInput struct {
	Method      string          `json:"method" validate:"required,oneof=cash debit_card credit_card e_wallet bank_transfer qris"`
	Amount      decimal.Decimal `json:"amount" validate:"required,gt=0"`
	ReferenceNo *string         `json:"reference_no"`
}

type PaymentResponse struct {
	ID             int64           `json:"id"`
	Method         string          `json:"method"`
	Amount         decimal.Decimal `json:"amount"`
	ReferenceNo    *string         `json:"reference_no"`
	Status         string          `json:"status"`
	RefundedAmount decimal.Decimal `json:"refunded_amount"`
	RefundedAt     *time.Time      `json:"refunded_at,omitempty"`
	RefundReason   *string         `json:"refund_reason,omitempty"`
	PaidAt         time.Time       `json:"paid_at"`
}

type RefundPaymentRequest struct {
	PaymentID    int64           `json:"payment_id" validate:"required"`
	Amount       decimal.Decimal `json:"amount" validate:"required,gt=0"`
	RefundReason string          `json:"refund_reason" validate:"required"`
}

// ============== Order Request DTOs ==============

type CreateOrderRequest struct {
	CustomerID      *int64           `json:"customer_id"`
	FulfillmentType string           `json:"fulfillment_type" validate:"omitempty,oneof=counter delivery"`
	ShippingAddress *string          `json:"shipping_address"`
	Items           []OrderItemInput `json:"items" validate:"required,min=1,dive"`
	Payments        []PaymentInput   `json:"payments"` // Optional - can add payments later
	Notes           *string          `json:"notes"`
	OfflineID       *uuid.UUID       `json:"offline_id"` // For offline sync
	PromoCode       *string          `json:"promo_code"`
}

type UpdateOrderRequest struct {
	CustomerID      *int64           `json:"customer_id"`
	FulfillmentType *string          `json:"fulfillment_type" validate:"omitempty,oneof=counter delivery"`
	ShippingAddress *string          `json:"shipping_address"`
	Items           []OrderItemInput `json:"items" validate:"omitempty,min=1,dive"`
	Notes           *string          `json:"notes"`
}

// Order lifecycle action requests
type ConfirmOrderRequest struct {
	// No additional fields needed for confirmation
}

type AddPaymentRequest struct {
	Payments []PaymentInput `json:"payments" validate:"required,min=1,dive"`
}

type CompleteOrderRequest struct {
	// For delivery orders, can optionally mark delivery status
	FulfillmentStatus *string `json:"fulfillment_status" validate:"omitempty,oneof=delivered"`
}

type CancelOrderRequest struct {
	Reason string `json:"reason" validate:"required"`
}

type VoidOrderRequest struct {
	Reason string `json:"reason" validate:"required"`
}

type ListOrderRequest struct {
	Page              int     `query:"page"`
	PerPage           int     `query:"per_page"`
	Search            string  `query:"search"`
	Status            *string `query:"status"`
	PaymentStatus     *string `query:"payment_status"`
	FulfillmentType   *string `query:"fulfillment_type"`
	FulfillmentStatus *string `query:"fulfillment_status"`
	CustomerID        *int64  `query:"customer_id"`
	PaymentMethod     *string `query:"payment_method"`
	DateFrom          *string `query:"date_from"`
	DateTo            *string `query:"date_to"`
}

// ============== Order Response DTOs ==============

type AppliedPromotionDTO struct {
	ID             int64           `json:"id"`
	Code           string          `json:"code"`
	Name           string          `json:"name"`
	DiscountAmount decimal.Decimal `json:"discount_amount"`
}

type OrderResponse struct {
	ID               int64                `json:"id"`
	OrderNo          string               `json:"order_no"`
	CustomerID       *int64               `json:"customer_id"`
	CustomerName     *string              `json:"customer_name,omitempty"`
	CashierID        int64                `json:"cashier_id"`
	CashierName      string               `json:"cashier_name,omitempty"`
	Status           string               `json:"status"`
	TotalAmount      decimal.Decimal      `json:"total_amount"`
	TotalDiscount    decimal.Decimal      `json:"total_discount"`
	TotalTax         decimal.Decimal      `json:"total_tax"`
	GrandTotal       decimal.Decimal      `json:"grand_total"`
	Notes            *string              `json:"notes"`
	AppliedPromotion *AppliedPromotionDTO `json:"applied_promotion,omitempty"`

	// Lifecycle fields
	PaymentStatus     string     `json:"payment_status"`
	FulfillmentType   string     `json:"fulfillment_type"`
	FulfillmentStatus *string    `json:"fulfillment_status,omitempty"`
	ShippingAddress   *string    `json:"shipping_address,omitempty"`
	ConfirmedAt       *time.Time `json:"confirmed_at,omitempty"`
	PaidAt            *time.Time `json:"paid_at,omitempty"`
	CompletedAt       *time.Time `json:"completed_at,omitempty"`
	CancelledAt       *time.Time `json:"cancelled_at,omitempty"`
	VoidedAt          *time.Time `json:"voided_at,omitempty"`
	CancelReason      *string    `json:"cancel_reason,omitempty"`
	VoidReason        *string    `json:"void_reason,omitempty"`

	// Offline sync
	OfflineID *uuid.UUID `json:"offline_id,omitempty"`
	SyncedAt  *time.Time `json:"synced_at,omitempty"`

	// Computed fields
	PaidAmount    decimal.Decimal `json:"paid_amount"`
	BalanceDue    decimal.Decimal `json:"balance_due"`
	RefundedTotal decimal.Decimal `json:"refunded_total"`

	Items     []*OrderItemResponse `json:"items"`
	Payments  []*PaymentResponse   `json:"payments"`
	CreatedAt time.Time            `json:"created_at"`
	UpdatedAt time.Time            `json:"updated_at"`
}

type OrderListResponse struct {
	Orders     []*OrderResponse `json:"orders"`
	Pagination *PaginationMeta  `json:"pagination"`
}

// ============== Order Preview DTOs ==============

type PreviewOrderRequest struct {
	Items     []OrderItemInput `json:"items" validate:"required,min=1,dive"`
	PromoCode *string          `json:"promo_code"`
}

type PreviewOrderResponse struct {
	Subtotal         decimal.Decimal      `json:"subtotal"`
	ItemDiscount     decimal.Decimal      `json:"item_discount"`
	PromoDiscount    decimal.Decimal      `json:"promo_discount"`
	Tax              decimal.Decimal      `json:"tax"`
	GrandTotal       decimal.Decimal      `json:"grand_total"`
	AppliedPromotion *AppliedPromotionDTO `json:"applied_promotion,omitempty"`
}
