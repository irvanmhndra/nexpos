package model

import "time"

type Company struct {
	ID        int64     `db:"id" json:"id"`
	Code      string    `db:"code" json:"code"`
	Name      string    `db:"name" json:"name"`
	Address   *string   `db:"address" json:"address,omitempty"`
	Phone     *string   `db:"phone" json:"phone,omitempty"`
	Email     *string   `db:"email" json:"email,omitempty"`
	TaxID     *string   `db:"tax_id" json:"tax_id,omitempty"`
	LogoURL   *string   `db:"logo_url" json:"logo_url,omitempty"`
	IsActive  bool      `db:"is_active" json:"is_active"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
