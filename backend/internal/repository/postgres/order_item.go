package postgres

import (
	"context"

	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/jmoiron/sqlx"
)

type OrderItemRepository struct {
	db *sqlx.DB
}

func NewOrderItemRepository(db *sqlx.DB) *OrderItemRepository {
	return &OrderItemRepository{db: db}
}

func (r *OrderItemRepository) Create(ctx context.Context, item *model.OrderItem) error {
	query := `
		INSERT INTO order_items (order_id, product_id, product_variant_id, sku, product_name,
			variant_name, variant_attributes, unit_price, unit_cost, quantity, discount_amount,
			tax_amount, subtotal, cogs_amount)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, created_at
	`
	return r.db.QueryRowContext(ctx, query,
		item.OrderID,
		item.ProductID,
		item.ProductVariantID,
		item.SKU,
		item.ProductName,
		item.VariantName,
		item.VariantAttributes,
		item.UnitPrice,
		item.UnitCost,
		item.Quantity,
		item.DiscountAmount,
		item.TaxAmount,
		item.Subtotal,
		item.CogsAmount,
	).Scan(&item.ID, &item.CreatedAt)
}

func (r *OrderItemRepository) GetByOrderID(ctx context.Context, orderID int64) ([]*model.OrderItem, error) {
	var items []*model.OrderItem
	query := `
		SELECT id, order_id, product_id, product_variant_id, sku, product_name,
			variant_name, variant_attributes, unit_price, unit_cost, quantity,
			discount_amount, tax_amount, subtotal, cogs_amount, created_at
		FROM order_items
		WHERE order_id = $1
		ORDER BY id
	`
	if err := r.db.SelectContext(ctx, &items, query, orderID); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *OrderItemRepository) DeleteByOrderID(ctx context.Context, orderID int64) error {
	query := `DELETE FROM order_items WHERE order_id = $1`
	_, err := r.db.ExecContext(ctx, query, orderID)
	return err
}
