package dto

import "time"

// ============== Requests ==============

type CreateCustomerRequest struct {
	Code     string  `json:"code" validate:"required,min=1,max=50"`
	Name     string  `json:"name" validate:"required,min=1,max=255"`
	Phone    *string `json:"phone" validate:"omitempty,max=50"`
	Email    *string `json:"email" validate:"omitempty,email,max=255"`
	IsMember bool    `json:"is_member"`
}

type UpdateCustomerRequest struct {
	Code     string  `json:"code" validate:"required,min=1,max=50"`
	Name     string  `json:"name" validate:"required,min=1,max=255"`
	Phone    *string `json:"phone" validate:"omitempty,max=50"`
	Email    *string `json:"email" validate:"omitempty,email,max=255"`
	IsMember bool    `json:"is_member"`
}

type ListCustomerRequest struct {
	Page     int    `query:"page"`
	PerPage  int    `query:"per_page"`
	Search   string `query:"search"`
	IsMember *bool  `query:"is_member"`
}

// ============== Responses ==============

type CustomerResponse struct {
	ID        int64     `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Phone     *string   `json:"phone"`
	Email     *string   `json:"email"`
	IsMember  bool      `json:"is_member"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CustomerListResponse struct {
	Customers  []*CustomerResponse `json:"customers"`
	Pagination *PaginationMeta     `json:"pagination"`
}

type PaginationMeta struct {
	TotalRecords int  `json:"total_records"`
	TotalPages   int  `json:"total_pages"`
	CurrentPage  int  `json:"current_page"`
	PerPage      int  `json:"per_page"`
	NextPage     *int `json:"next_page"`
	PrevPage     *int `json:"prev_page"`
}
