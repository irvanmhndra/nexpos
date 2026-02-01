package model

import "time"

type Product struct {
	ID                int64     `db:"id" json:"id"`
	CompanyID         int64     `db:"company_id" json:"company_id"`
	ProductCategoryID *int64    `db:"product_category_id" json:"product_category_id"`
	Name              string    `db:"name" json:"name"`
	Description       *string   `db:"description" json:"description"`
	ImageData         *string   `db:"image_data" json:"image_data"`
	IsActive          bool      `db:"is_active" json:"is_active"`
	CreatedAt         time.Time `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time `db:"updated_at" json:"updated_at"`

	// Eager loaded (not from DB)
	Category *ProductCategory  `db:"-" json:"-"`
	Variants []*ProductVariant `db:"-" json:"-"`
}
