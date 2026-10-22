package model

import (
	"time"
)

type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
)

type User struct {
	ID           int64      `db:"id" json:"id"`
	CompanyID    int64      `db:"company_id" json:"company_id"`
	RoleID       *int64     `db:"role_id" json:"role_id"`
	Email        string     `db:"email" json:"email"`
	PasswordHash string     `db:"password_hash" json:"-"`
	Name         string     `db:"name" json:"name"`
	Status       UserStatus `db:"status" json:"status"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`

	// Eager loaded
	Role     *Role    `db:"-" json:"-"`
	Branches []Branch `db:"-" json:"-"`
}

func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

// HasPermission checks if user has a specific permission through their role
func (u *User) HasPermission(code string) bool {
	if u.Role == nil {
		return false
	}
	return u.Role.HasPermission(code)
}

// IsOwner checks if user has owner role
func (u *User) IsOwner() bool {
	if u.Role == nil {
		return false
	}
	return u.Role.IsOwner()
}
