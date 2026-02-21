package dto

import "time"

// ============== Requests ==============

type CreatePromotionRequest struct {
	Code          string   `json:"code" validate:"required,min=1,max=50"`
	Name          string   `json:"name" validate:"required,min=1,max=255"`
	Description   *string  `json:"description" validate:"omitempty,max=1000"`
	Type          string   `json:"type" validate:"required,oneof=discount bundle conditional"`
	DiscountType  *string  `json:"discount_type" validate:"omitempty,oneof=percentage fixed"`
	DiscountValue *float64 `json:"discount_value" validate:"omitempty,gte=0"`
	MinPurchase   *float64 `json:"min_purchase" validate:"omitempty,gte=0"`
	MaxDiscount   *float64 `json:"max_discount" validate:"omitempty,gte=0"`
	StartAt       string   `json:"start_at" validate:"required"` // ISO 8601
	EndAt         *string  `json:"end_at" validate:"omitempty"`  // ISO 8601
	Priority      int      `json:"priority" validate:"gte=0"`
	IsActive      bool     `json:"is_active"`
}

type UpdatePromotionRequest struct {
	Code          string   `json:"code" validate:"required,min=1,max=50"`
	Name          string   `json:"name" validate:"required,min=1,max=255"`
	Description   *string  `json:"description" validate:"omitempty,max=1000"`
	Type          string   `json:"type" validate:"required,oneof=discount bundle conditional"`
	DiscountType  *string  `json:"discount_type" validate:"omitempty,oneof=percentage fixed"`
	DiscountValue *float64 `json:"discount_value" validate:"omitempty,gte=0"`
	MinPurchase   *float64 `json:"min_purchase" validate:"omitempty,gte=0"`
	MaxDiscount   *float64 `json:"max_discount" validate:"omitempty,gte=0"`
	StartAt       string   `json:"start_at" validate:"required"`
	EndAt         *string  `json:"end_at" validate:"omitempty"`
	Priority      int      `json:"priority" validate:"gte=0"`
	IsActive      bool     `json:"is_active"`
}

type ListPromotionRequest struct {
	Page     int     `query:"page"`
	PerPage  int     `query:"per_page"`
	Search   string  `query:"search"`
	IsActive *bool   `query:"is_active"`
	Type     string  `query:"type"`
}

// ============== Responses ==============

type PromotionResponse struct {
	ID            int64      `json:"id"`
	Code          string     `json:"code"`
	Name          string     `json:"name"`
	Description   *string    `json:"description"`
	Type          string     `json:"type"`
	DiscountType  *string    `json:"discount_type"`
	DiscountValue *float64   `json:"discount_value"`
	MinPurchase   *float64   `json:"min_purchase"`
	MaxDiscount   *float64   `json:"max_discount"`
	StartAt       time.Time  `json:"start_at"`
	EndAt         *time.Time `json:"end_at"`
	Priority      int        `json:"priority"`
	IsActive      bool       `json:"is_active"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type PromotionListResponse struct {
	Promotions []*PromotionResponse `json:"promotions"`
	Pagination *PaginationMeta      `json:"pagination"`
}
