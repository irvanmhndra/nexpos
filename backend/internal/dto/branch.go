package dto

import "time"

// ============== Requests ==============

type CreateBranchRequest struct {
	Code     string  `json:"code" validate:"required,min=1,max=50"`
	Name     string  `json:"name" validate:"required,min=1,max=255"`
	Address  *string `json:"address" validate:"omitempty,max=500"`
	Phone    *string `json:"phone" validate:"omitempty,max=50"`
	IsActive bool    `json:"is_active"`
}

type UpdateBranchRequest struct {
	Code     string  `json:"code" validate:"required,min=1,max=50"`
	Name     string  `json:"name" validate:"required,min=1,max=255"`
	Address  *string `json:"address" validate:"omitempty,max=500"`
	Phone    *string `json:"phone" validate:"omitempty,max=50"`
	IsActive bool    `json:"is_active"`
}

type ListBranchRequest struct {
	Page     int    `query:"page"`
	PerPage  int    `query:"per_page"`
	Search   string `query:"search"`
	IsActive *bool  `query:"is_active"`
}

// ============== Responses ==============

type BranchResponse struct {
	ID        int64     `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Address   *string   `json:"address"`
	Phone     *string   `json:"phone"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type BranchListResponse struct {
	Branches   []*BranchResponse `json:"branches"`
	Pagination *PaginationMeta   `json:"pagination"`
}
