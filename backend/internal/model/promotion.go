package model

import "time"

type Promotion struct {
	ID            int64      `db:"id" json:"id"`
	CompanyID     int64      `db:"company_id" json:"company_id"`
	Code          string     `db:"code" json:"code"`
	Name          string     `db:"name" json:"name"`
	Description   *string    `db:"description" json:"description,omitempty"`
	Type          string     `db:"type" json:"type"` // discount | bundle | conditional
	DiscountType  *string    `db:"discount_type" json:"discount_type,omitempty"` // percentage | fixed
	DiscountValue *float64   `db:"discount_value" json:"discount_value,omitempty"`
	MinPurchase   *float64   `db:"min_purchase" json:"min_purchase,omitempty"`
	MaxDiscount   *float64   `db:"max_discount" json:"max_discount,omitempty"`
	StartAt       time.Time  `db:"start_at" json:"start_at"`
	EndAt         *time.Time `db:"end_at" json:"end_at,omitempty"`
	Priority      int        `db:"priority" json:"priority"`
	IsActive      bool       `db:"is_active" json:"is_active"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at" json:"updated_at"`
}

// Promotion lifecycle status, derived from is_active + the schedule (never stored,
// so it can't drift). Evaluated against server time.
const (
	PromotionStatusInactive  = "inactive"  // manually turned off
	PromotionStatusScheduled = "scheduled" // active but not started yet
	PromotionStatusActive    = "active"    // active and within the window
	PromotionStatusExpired   = "expired"   // active but past end_at
)

// Status returns the effective status at the given (server) time.
func (p *Promotion) Status(now time.Time) string {
	if !p.IsActive {
		return PromotionStatusInactive
	}
	if p.StartAt.After(now) {
		return PromotionStatusScheduled
	}
	if p.EndAt != nil && p.EndAt.Before(now) {
		return PromotionStatusExpired
	}
	return PromotionStatusActive
}
