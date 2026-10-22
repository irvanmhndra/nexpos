package dto

import "time"

// ============== Requests ==============

type CreateProductCategoryRequest struct {
	Code      string `json:"code" validate:"required,min=1,max=50"`
	Name      string `json:"name" validate:"required,min=1,max=255"`
	ParentID  *int64 `json:"parent_id" validate:"omitempty"`
	SortOrder int    `json:"sort_order" validate:"gte=0"`
	IsActive  *bool  `json:"is_active"`
}

type UpdateProductCategoryRequest struct {
	Code      string `json:"code" validate:"required,min=1,max=50"`
	Name      string `json:"name" validate:"required,min=1,max=255"`
	ParentID  *int64 `json:"parent_id" validate:"omitempty"`
	SortOrder int    `json:"sort_order" validate:"gte=0"`
	IsActive  *bool  `json:"is_active"`
}

type ListProductCategoryRequest struct {
	Page     int    `query:"page"`
	PerPage  int    `query:"per_page"`
	Search   string `query:"search"`
	IsActive *bool  `query:"is_active"`
}

// ============== Responses ==============

type ProductCategoryResponse struct {
	ID         int64     `json:"id"`
	Code       string    `json:"code"`
	Name       string    `json:"name"`
	ParentID   *int64    `json:"parent_id"`
	ParentName *string   `json:"parent_name,omitempty"`
	SortOrder  int       `json:"sort_order"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type ProductCategoryListResponse struct {
	Categories []*ProductCategoryResponse `json:"categories"`
	Pagination *PaginationMeta            `json:"pagination"`
}
