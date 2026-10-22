package model

import "time"

type Stock struct {
	ID               int64     `db:"id"                 json:"id"`
	ProductVariantID int64     `db:"product_variant_id" json:"product_variant_id"`
	BranchID         int64     `db:"branch_id"          json:"branch_id"`
	Quantity         int       `db:"quantity"           json:"quantity"`
	MinQuantity      int       `db:"min_quantity"       json:"min_quantity"`
	UpdatedAt        time.Time `db:"updated_at"         json:"updated_at"`
}
