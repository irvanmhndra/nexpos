package model

import "time"

type Role struct {
	ID          int64     `db:"id"`
	CompanyID   *int64    `db:"company_id"` // NULL = system role
	Code        string    `db:"code"`
	Name        string    `db:"name"`
	Description *string   `db:"description"`
	IsSystem    bool      `db:"is_system"`
	IsActive    bool      `db:"is_active"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`

	// Eager loaded
	Permissions []Permission `db:"-"`
}

// IsOwner checks if this is the owner role
func (r *Role) IsOwner() bool {
	return r.Code == "owner"
}

// HasPermission checks if role has a specific permission
func (r *Role) HasPermission(code string) bool {
	for _, p := range r.Permissions {
		if p.Code == code {
			return true
		}
	}
	return false
}
