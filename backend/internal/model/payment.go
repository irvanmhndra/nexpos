package model

import "time"

type Payment struct {
	ID          int64     `db:"id" json:"id"`
	OrderID     int64     `db:"order_id" json:"order_id"`
	Method      string    `db:"method" json:"method"`
	Amount      float64   `db:"amount" json:"amount"`
	ReferenceNo *string   `db:"reference_no" json:"reference_no"`
	PaidAt      time.Time `db:"paid_at" json:"paid_at"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`

	// Refund support
	Status         string     `db:"status" json:"status"`
	RefundedAmount float64    `db:"refunded_amount" json:"refunded_amount"`
	RefundedAt     *time.Time `db:"refunded_at" json:"refunded_at"`
	RefundReason   *string    `db:"refund_reason" json:"refund_reason"`
}

// Payment method constants
const (
	PaymentMethodCash         = "cash"
	PaymentMethodDebitCard    = "debit_card"
	PaymentMethodCreditCard   = "credit_card"
	PaymentMethodEWallet      = "e_wallet"
	PaymentMethodBankTransfer = "bank_transfer"
	PaymentMethodQRIS         = "qris"
)

// Payment status constants
const (
	PaymentStatusPending           = "pending"
	PaymentStatusCompleted         = "completed"
	PaymentStatusRefunded          = "refunded"
	PaymentStatusPartiallyRefunded = "partially_refunded"
	PaymentStatusFailed            = "failed"
)
