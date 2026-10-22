package model

import "time"

type ProductCategory struct {
	ID        int64     `db:"id" json:"id"`
	CompanyID int64     `db:"company_id" json:"company_id"`
	ParentID  *int64    `db:"parent_id" json:"parent_id"`
	Code      string    `db:"code" json:"code"`
	Name      string    `db:"name" json:"name"`
	SortOrder int       `db:"sort_order" json:"sort_order"`
	IsActive  bool      `db:"is_active" json:"is_active"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`

	// Eager loaded (not from DB)
	Parent *ProductCategory `db:"-" json:"-"`
}
