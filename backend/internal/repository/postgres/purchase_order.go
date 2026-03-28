package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/irvanmhndra/nexpos-api/internal/repository"
	"github.com/jmoiron/sqlx"
)

type PurchaseOrderRepository struct {
	db *sqlx.DB
}

func NewPurchaseOrderRepository(db *sqlx.DB) *PurchaseOrderRepository {
	return &PurchaseOrderRepository{db: db}
}

func (r *PurchaseOrderRepository) Create(ctx context.Context, po *model.PurchaseOrder) error {
	query := `
		INSERT INTO purchase_orders (company_id, branch_id, supplier_id, po_number, status, notes, total_amount, ordered_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		po.CompanyID,
		po.BranchID,
		po.SupplierID,
		po.PONumber,
		po.Status,
		po.Notes,
		po.TotalAmount,
		po.OrderedAt,
	).Scan(&po.ID, &po.CreatedAt, &po.UpdatedAt)
}

func (r *PurchaseOrderRepository) GetByID(ctx context.Context, companyID, id int64) (*model.PurchaseOrder, error) {
	var po model.PurchaseOrder
	query := `SELECT * FROM purchase_orders WHERE id = $1 AND company_id = $2`
	err := r.db.GetContext(ctx, &po, query, id, companyID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &po, nil
}

func (r *PurchaseOrderRepository) GeneratePONumber(ctx context.Context, companyID int64) (string, error) {
	now := time.Now()
	prefix := fmt.Sprintf("PO-%d%02d%02d-", now.Year(), now.Month(), now.Day())

	var count int
	query := `SELECT COUNT(*) FROM purchase_orders WHERE company_id = $1 AND po_number LIKE $2`
	if err := r.db.GetContext(ctx, &count, query, companyID, prefix+"%"); err != nil {
		return "", err
	}

	return fmt.Sprintf("%s%04d", prefix, count+1), nil
}

func (r *PurchaseOrderRepository) List(ctx context.Context, companyID int64, params *repository.POListParams) ([]*model.PurchaseOrder, int, error) {
	var pos []*model.PurchaseOrder
	var total int

	whereClause := "WHERE company_id = $1"
	args := []interface{}{companyID}
	argIndex := 2

	if params.Status != "" {
		whereClause += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, params.Status)
		argIndex++
	}

	if params.SupplierID != nil {
		whereClause += fmt.Sprintf(" AND supplier_id = $%d", argIndex)
		args = append(args, *params.SupplierID)
		argIndex++
	}

	countQuery := "SELECT COUNT(*) FROM purchase_orders " + whereClause
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	listQuery := fmt.Sprintf(`
		SELECT * FROM purchase_orders %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)
	args = append(args, params.Limit, params.Offset)

	if err := r.db.SelectContext(ctx, &pos, listQuery, args...); err != nil {
		return nil, 0, err
	}

	return pos, total, nil
}

func (r *PurchaseOrderRepository) Update(ctx context.Context, po *model.PurchaseOrder) error {
	query := `
		UPDATE purchase_orders
		SET status = $1, notes = $2, total_amount = $3, ordered_at = $4, received_at = $5, cancelled_at = $6, updated_at = NOW()
		WHERE id = $7 AND company_id = $8
	`
	result, err := r.db.ExecContext(ctx, query,
		po.Status,
		po.Notes,
		po.TotalAmount,
		po.OrderedAt,
		po.ReceivedAt,
		po.CancelledAt,
		po.ID,
		po.CompanyID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *PurchaseOrderRepository) GetItems(ctx context.Context, poID int64) ([]*model.PurchaseOrderItem, error) {
	var items []*model.PurchaseOrderItem
	query := `SELECT * FROM purchase_order_items WHERE purchase_order_id = $1 ORDER BY id`
	if err := r.db.SelectContext(ctx, &items, query, poID); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *PurchaseOrderRepository) CreateItem(ctx context.Context, item *model.PurchaseOrderItem) error {
	query := `
		INSERT INTO purchase_order_items (purchase_order_id, product_variant_id, sku, variant_name, quantity, received_quantity, unit_cost, subtotal)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		item.PurchaseOrderID,
		item.ProductVariantID,
		item.SKU,
		item.VariantName,
		item.Quantity,
		item.ReceivedQuantity,
		item.UnitCost,
		item.Subtotal,
	).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)
}

func (r *PurchaseOrderRepository) UpdateItem(ctx context.Context, item *model.PurchaseOrderItem) error {
	query := `
		UPDATE purchase_order_items
		SET received_quantity = $1, updated_at = NOW()
		WHERE id = $2
	`
	_, err := r.db.ExecContext(ctx, query, item.ReceivedQuantity, item.ID)
	return err
}
