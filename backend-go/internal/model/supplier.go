package model

import "time"

type Supplier struct {
	ID          int64     `db:"id"           json:"id"`
	CompanyID   int64     `db:"company_id"   json:"company_id"`
	Name        string    `db:"name"         json:"name"`
	Code        string    `db:"code"         json:"code"`
	ContactName *string   `db:"contact_name" json:"contact_name"`
	Phone       *string   `db:"phone"        json:"phone"`
	Email       *string   `db:"email"        json:"email"`
	Address     *string   `db:"address"      json:"address"`
	Notes       *string   `db:"notes"        json:"notes"`
	IsActive    bool      `db:"is_active"    json:"is_active"`
	CreatedAt   time.Time `db:"created_at"   json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"   json:"updated_at"`
}
