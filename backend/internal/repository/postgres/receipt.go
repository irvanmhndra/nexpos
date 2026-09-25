package postgres

import (
	"context"
	"database/sql"

	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/jmoiron/sqlx"
)

type ReceiptRepository struct {
	db *sqlx.DB
}

func NewReceiptRepository(db *sqlx.DB) *ReceiptRepository {
	return &ReceiptRepository{db: db}
}

func (r *ReceiptRepository) Save(ctx context.Context, rec *model.Receipt) error {
	query := `
		INSERT INTO receipts (
			company_id, order_id, order_no, cashier_id, cashier_name,
			customer_id, customer_name, items, payments,
			total_amount, total_discount, total_tax, grand_total, notes,
			completed_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		ON CONFLICT (company_id, order_id) DO NOTHING
		RETURNING id, created_at
	`
	err := conn(ctx, r.db).QueryRowContext(ctx, query,
		rec.CompanyID,
		rec.OrderID,
		rec.OrderNo,
		rec.CashierID,
		rec.CashierName,
		rec.CustomerID,
		rec.CustomerName,
		rec.Items,
		rec.Payments,
		rec.TotalAmount,
		rec.TotalDiscount,
		rec.TotalTax,
		rec.GrandTotal,
		rec.Notes,
		rec.CompletedAt,
	).Scan(&rec.ID, &rec.CreatedAt)

	// ON CONFLICT with no RETURNING row → sql.ErrNoRows. Receipts are write-once;
	// a duplicate is a no-op, not an error.
	if err == sql.ErrNoRows {
		return nil
	}
	return err
}

func (r *ReceiptRepository) GetByOrderID(ctx context.Context, companyID, orderID int64) (*model.Receipt, error) {
	var rec model.Receipt
	query := `SELECT * FROM receipts WHERE company_id = $1 AND order_id = $2`
	err := conn(ctx, r.db).GetContext(ctx, &rec, query, companyID, orderID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &rec, nil
}
