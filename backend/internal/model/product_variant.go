package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

type ProductVariant struct {
	ID               int64     `db:"id" json:"id"`
	ProductID        int64     `db:"product_id" json:"product_id"`
	SKU              string    `db:"sku" json:"sku"`
	Name             string    `db:"name" json:"name"`
	Attributes       JSONMap   `db:"attributes" json:"attributes"`
	Price            float64   `db:"price" json:"price"`
	StandardCost     float64   `db:"standard_cost" json:"standard_cost"`
	LastPurchaseCost float64   `db:"last_purchase_cost" json:"last_purchase_cost"`
	IsDefault        bool      `db:"is_default" json:"is_default"`
	IsActive         bool      `db:"is_active" json:"is_active"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time `db:"updated_at" json:"updated_at"`
}

// JSONMap is a custom type for handling JSONB columns
type JSONMap map[string]interface{}

func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}
