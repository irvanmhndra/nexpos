package dto

import (
	"time"

	"github.com/irvanmhndra/nexpos-api/pkg/httputil"
)

// ============== Requests ==============

type CreateSupplierRequest struct {
	Name        string  `json:"name"         validate:"required,min=1,max=200"`
	Code        string  `json:"code"         validate:"required,min=1,max=50"`
	ContactName *string `json:"contact_name" validate:"omitempty,max=100"`
	Phone       *string `json:"phone"        validate:"omitempty,max=50"`
	Email       *string `json:"email"        validate:"omitempty,email,max=150"`
	Address     *string `json:"address"`
	Notes       *string `json:"notes"`
	IsActive    bool    `json:"is_active"`
}

type UpdateSupplierRequest struct {
	Name        string  `json:"name"         validate:"required,min=1,max=200"`
	Code        string  `json:"code"         validate:"required,min=1,max=50"`
	ContactName *string `json:"contact_name" validate:"omitempty,max=100"`
	Phone       *string `json:"phone"        validate:"omitempty,max=50"`
	Email       *string `json:"email"        validate:"omitempty,email,max=150"`
	Address     *string `json:"address"`
	Notes       *string `json:"notes"`
	IsActive    bool    `json:"is_active"`
}

type ListSupplierRequest struct {
	Page     int    `query:"page"`
	PerPage  int    `query:"per_page"`
	Search   string `query:"search"`
	IsActive *bool  `query:"is_active"`
}

// ============== Responses ==============

type SupplierResponse struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	ContactName *string   `json:"contact_name"`
	Phone       *string   `json:"phone"`
	Email       *string   `json:"email"`
	Address     *string   `json:"address"`
	Notes       *string   `json:"notes"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type SupplierListResponse struct {
	Suppliers  []*SupplierResponse  `json:"suppliers"`
	Pagination *httputil.Pagination `json:"pagination"`
}
