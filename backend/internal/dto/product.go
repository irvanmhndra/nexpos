package dto

import "time"

// ============== Variant DTOs ==============

type ProductVariantInput struct {
	ID               *int64                 `json:"id"`
	SKU              string                 `json:"sku" validate:"required,min=1,max=100"`
	Barcode          *string                `json:"barcode" validate:"omitempty,max=100"`
	Name             string                 `json:"name" validate:"required,min=1,max=255"`
	Attributes       map[string]interface{} `json:"attributes"`
	Price            float64                `json:"price" validate:"gte=0"`
	StandardCost     float64                `json:"standard_cost" validate:"gte=0"`
	LastPurchaseCost float64                `json:"last_purchase_cost" validate:"gte=0"`
	IsDefault        bool                   `json:"is_default"`
	IsActive         *bool                  `json:"is_active"`
	SalePrice        *float64               `json:"sale_price"`
	SaleStart        *string                `json:"sale_start"`
	SaleEnd          *string                `json:"sale_end"`
}

type ProductVariantResponse struct {
	ID               int64                  `json:"id"`
	SKU              string                 `json:"sku"`
	Barcode          *string                `json:"barcode,omitempty"`
	Name             string                 `json:"name"`
	Attributes       map[string]interface{} `json:"attributes"`
	Price            float64                `json:"price"`
	StandardCost     float64                `json:"standard_cost"`
	LastPurchaseCost float64                `json:"last_purchase_cost"`
	IsDefault        bool                   `json:"is_default"`
	IsActive         bool                   `json:"is_active"`
	SalePrice        *float64               `json:"sale_price,omitempty"`
	SaleStart        *time.Time             `json:"sale_start,omitempty"`
	SaleEnd          *time.Time             `json:"sale_end,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

// ============== Product Requests ==============

type CreateProductRequest struct {
	Name        string                `json:"name" validate:"required,min=1,max=255"`
	CategoryID  *int64                `json:"category_id"`
	Description *string               `json:"description"`
	ImageData   *string               `json:"image_data"`
	ImageURL    *string               `json:"image_url"`
	IsActive      *bool                 `json:"is_active"`
	Variants      []ProductVariantInput `json:"variants" validate:"required,min=1,dive"`
}

type UpdateProductRequest struct {
	Name        string                `json:"name" validate:"required,min=1,max=255"`
	CategoryID  *int64                `json:"category_id"`
	Description *string               `json:"description"`
	ImageData   *string               `json:"image_data"`
	ImageURL    *string               `json:"image_url"`
	IsActive      *bool                 `json:"is_active"`
	Variants      []ProductVariantInput `json:"variants" validate:"required,min=1,dive"`
}

type ListProductRequest struct {
	Page       int    `query:"page"`
	PerPage    int    `query:"per_page"`
	Search     string `query:"search"`
	CategoryID *int64 `query:"category_id"`
	IsActive   *bool  `query:"is_active"`
}

// ============== Product Responses ==============

type ProductResponse struct {
	ID           int64                     `json:"id"`
	Name         string                    `json:"name"`
	CategoryID   *int64                    `json:"category_id"`
	CategoryName *string                   `json:"category_name,omitempty"`
	Description  *string                   `json:"description"`
	ImageData    *string                   `json:"image_data"`
	ImageURL      *string                  `json:"image_url"`
	IsActive      bool                     `json:"is_active"`
	Variants      []*ProductVariantResponse `json:"variants"`
	CreatedAt    time.Time                 `json:"created_at"`
	UpdatedAt    time.Time                 `json:"updated_at"`
}

type ProductListResponse struct {
	Products   []*ProductResponse `json:"products"`
	Pagination *PaginationMeta    `json:"pagination"`
}

// ProductLookupResponse is returned when scanning a code at the POS: the owning
// product plus which variant matched and whether it matched on barcode or SKU.
type ProductLookupResponse struct {
	Product          *ProductResponse `json:"product"`
	MatchedVariantID int64            `json:"matched_variant_id"`
	MatchedBy        string           `json:"matched_by"` // "barcode" | "sku"
}
