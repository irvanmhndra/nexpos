package model

import "time"

type Branch struct {
	ID        int64     `db:"id" json:"id"`
	CompanyID int64     `db:"company_id" json:"company_id"`
	Code      string    `db:"code" json:"code"`
	Name      string    `db:"name" json:"name"`
	Address   *string   `db:"address" json:"address,omitempty"`
	Phone     *string   `db:"phone" json:"phone,omitempty"`
	IsActive  bool      `db:"is_active" json:"is_active"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
