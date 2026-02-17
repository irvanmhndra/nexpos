package model

import (
	"time"

	"github.com/google/uuid"
)

type Order struct {
	ID                int64     `db:"id" json:"id"`
	CompanyID         int64     `db:"company_id" json:"company_id"`
	BranchID          int64     `db:"branch_id" json:"branch_id"`
	OrderNo           string    `db:"order_no" json:"order_no"`
	CustomerID        *int64    `db:"customer_id" json:"customer_id"`
	CashierID         int64     `db:"cashier_id" json:"cashier_id"`
	Status            string    `db:"status" json:"status"`
	TotalAmount       float64   `db:"total_amount" json:"total_amount"`
	TotalDiscount     float64   `db:"total_discount" json:"total_discount"`
	TotalTax          float64   `db:"total_tax" json:"total_tax"`
	GrandTotal        float64   `db:"grand_total" json:"grand_total"`
	AppliedPromotions JSONMap   `db:"applied_promotions" json:"applied_promotions"`
	Notes             *string   `db:"notes" json:"notes"`
	CreatedAt         time.Time `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time `db:"updated_at" json:"updated_at"`

	// New lifecycle columns
	PaymentStatus     string     `db:"payment_status" json:"payment_status"`
	FulfillmentType   string     `db:"fulfillment_type" json:"fulfillment_type"`
	FulfillmentStatus *string    `db:"fulfillment_status" json:"fulfillment_status"`
	ShippingAddress   *string    `db:"shipping_address" json:"shipping_address"`
	ConfirmedAt       *time.Time `db:"confirmed_at" json:"confirmed_at"`
	PaidAt            *time.Time `db:"paid_at" json:"paid_at"`
	CompletedAt       *time.Time `db:"completed_at" json:"completed_at"`
	CancelledAt       *time.Time `db:"cancelled_at" json:"cancelled_at"`
	VoidedAt          *time.Time `db:"voided_at" json:"voided_at"`
	CancelReason      *string    `db:"cancel_reason" json:"cancel_reason"`
	VoidReason        *string    `db:"void_reason" json:"void_reason"`

	// Offline sync support
	OfflineID *uuid.UUID `db:"offline_id" json:"offline_id"`
	SyncedAt  *time.Time `db:"synced_at" json:"synced_at"`

	// Eager loaded
	Items    []*OrderItem `db:"-" json:"-"`
	Payments []*Payment   `db:"-" json:"-"`
}

// Order status constants (lifecycle: draft → confirmed → completed | cancelled | voided)
const (
	OrderStatusDraft     = "draft"
	OrderStatusConfirmed = "confirmed"
	OrderStatusCompleted = "completed"
	OrderStatusCancelled = "cancelled"
	OrderStatusVoided    = "voided"
)

// Order payment status constants (lifecycle: unpaid → partial → paid | refunded)
const (
	PaymentStatusUnpaid  = "unpaid"
	PaymentStatusPartial = "partial"
	PaymentStatusPaid    = "paid"
	// PaymentStatusRefunded is defined in payment.go
)

// Fulfillment type constants
const (
	FulfillmentTypeCounter  = "counter"
	FulfillmentTypeDelivery = "delivery"
)

// Fulfillment status constants (only for delivery orders)
const (
	FulfillmentStatusPending    = "pending"
	FulfillmentStatusProcessing = "processing"
	FulfillmentStatusShipped    = "shipped"
	FulfillmentStatusDelivered  = "delivered"
)
