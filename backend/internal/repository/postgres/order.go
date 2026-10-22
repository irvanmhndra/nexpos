package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/irvanmhndra/nexpos-api/internal/repository"
	"github.com/jmoiron/sqlx"
)

type OrderRepository struct {
	db *sqlx.DB
}

func NewOrderRepository(db *sqlx.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// All order columns for SELECT queries
const orderColumns = `id, company_id, branch_id, order_no, customer_id, cashier_id, status,
	total_amount, total_discount, total_tax, grand_total, applied_promotions, notes,
	payment_status, fulfillment_type, fulfillment_status, shipping_address,
	confirmed_at, paid_at, completed_at, cancelled_at, voided_at,
	cancel_reason, void_reason, offline_id, synced_at, created_at, updated_at`

func (r *OrderRepository) Create(ctx context.Context, order *model.Order) error {
	query := `
		INSERT INTO orders (company_id, branch_id, order_no, customer_id, cashier_id, status,
			total_amount, total_discount, total_tax, grand_total, applied_promotions, notes,
			payment_status, fulfillment_type, fulfillment_status, shipping_address, offline_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		RETURNING id, created_at, updated_at
	`
	return conn(ctx, r.db).QueryRowContext(ctx, query,
		order.CompanyID,
		order.BranchID,
		order.OrderNo,
		order.CustomerID,
		order.CashierID,
		order.Status,
		order.TotalAmount,
		order.TotalDiscount,
		order.TotalTax,
		order.GrandTotal,
		order.AppliedPromotions,
		order.Notes,
		order.PaymentStatus,
		order.FulfillmentType,
		order.FulfillmentStatus,
		order.ShippingAddress,
		order.OfflineID,
	).Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)
}

func (r *OrderRepository) GetByID(ctx context.Context, companyID, id int64) (*model.Order, error) {
	var order model.Order
	query := fmt.Sprintf(`SELECT %s FROM orders WHERE id = $1 AND company_id = $2`, orderColumns)
	err := conn(ctx, r.db).GetContext(ctx, &order, query, id, companyID)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepository) GetByIDForUpdate(ctx context.Context, companyID, id int64) (*model.Order, error) {
	if err := requireTx(ctx); err != nil {
		return nil, err
	}
	var order model.Order
	query := fmt.Sprintf(`SELECT %s FROM orders WHERE id = $1 AND company_id = $2 FOR UPDATE`, orderColumns)
	if err := conn(ctx, r.db).GetContext(ctx, &order, query, id, companyID); err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepository) GetByOrderNo(ctx context.Context, companyID int64, orderNo string) (*model.Order, error) {
	var order model.Order
	query := fmt.Sprintf(`SELECT %s FROM orders WHERE order_no = $1 AND company_id = $2`, orderColumns)
	err := conn(ctx, r.db).GetContext(ctx, &order, query, orderNo, companyID)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepository) GetByOfflineID(ctx context.Context, companyID int64, offlineID string) (*model.Order, error) {
	var order model.Order
	query := fmt.Sprintf(`SELECT %s FROM orders WHERE offline_id = $1 AND company_id = $2`, orderColumns)
	err := conn(ctx, r.db).GetContext(ctx, &order, query, offlineID, companyID)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepository) List(ctx context.Context, companyID int64, params *repository.OrderListParams) ([]*model.Order, int, error) {
	conditions := []string{"company_id = $1"}
	args := []interface{}{companyID}
	argPos := 2

	if params.Search != "" {
		conditions = append(conditions, fmt.Sprintf("order_no ILIKE $%d", argPos))
		args = append(args, "%"+params.Search+"%")
		argPos++
	}

	if params.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argPos))
		args = append(args, *params.Status)
		argPos++
	}

	if params.PaymentStatus != nil {
		conditions = append(conditions, fmt.Sprintf("payment_status = $%d", argPos))
		args = append(args, *params.PaymentStatus)
		argPos++
	}

	if params.FulfillmentType != nil {
		conditions = append(conditions, fmt.Sprintf("fulfillment_type = $%d", argPos))
		args = append(args, *params.FulfillmentType)
		argPos++
	}

	if params.FulfillmentStatus != nil {
		conditions = append(conditions, fmt.Sprintf("fulfillment_status = $%d", argPos))
		args = append(args, *params.FulfillmentStatus)
		argPos++
	}

	if params.CustomerID != nil {
		conditions = append(conditions, fmt.Sprintf("customer_id = $%d", argPos))
		args = append(args, *params.CustomerID)
		argPos++
	}

	if params.DateFrom != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argPos))
		args = append(args, *params.DateFrom)
		argPos++
	}

	if params.DateTo != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d::date + interval '1 day'", argPos))
		args = append(args, *params.DateTo)
		argPos++
	}

	whereClause := strings.Join(conditions, " AND ")

	// Count query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM orders WHERE %s", whereClause)
	var total int
	if err := conn(ctx, r.db).GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	// Data query
	dataQuery := fmt.Sprintf(`
		SELECT %s
		FROM orders
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, orderColumns, whereClause, argPos, argPos+1)

	args = append(args, params.Limit, params.Offset)

	var orders []*model.Order
	if err := conn(ctx, r.db).SelectContext(ctx, &orders, dataQuery, args...); err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

func (r *OrderRepository) Update(ctx context.Context, order *model.Order) error {
	query := `
		UPDATE orders
		SET customer_id = $1, status = $2, total_amount = $3, total_discount = $4, total_tax = $5,
			grand_total = $6, notes = $7, payment_status = $8, fulfillment_type = $9,
			fulfillment_status = $10, shipping_address = $11, confirmed_at = $12, paid_at = $13,
			completed_at = $14, cancelled_at = $15, voided_at = $16, cancel_reason = $17,
			void_reason = $18, synced_at = $19, updated_at = NOW()
		WHERE id = $20 AND company_id = $21
	`
	result, err := conn(ctx, r.db).ExecContext(ctx, query,
		order.CustomerID,
		order.Status,
		order.TotalAmount,
		order.TotalDiscount,
		order.TotalTax,
		order.GrandTotal,
		order.Notes,
		order.PaymentStatus,
		order.FulfillmentType,
		order.FulfillmentStatus,
		order.ShippingAddress,
		order.ConfirmedAt,
		order.PaidAt,
		order.CompletedAt,
		order.CancelledAt,
		order.VoidedAt,
		order.CancelReason,
		order.VoidReason,
		order.SyncedAt,
		order.ID,
		order.CompanyID,
	)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("order not found")
	}

	return nil
}

func (r *OrderRepository) Delete(ctx context.Context, companyID, id int64) error {
	query := `DELETE FROM orders WHERE id = $1 AND company_id = $2`
	result, err := conn(ctx, r.db).ExecContext(ctx, query, id, companyID)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("order not found")
	}

	return nil
}

func (r *OrderRepository) GenerateOrderNo(ctx context.Context, companyID, branchID int64) (string, error) {
	// Format: ORD-YYYYMMDD-XXXX
	// Example: ORD-20260201-0001
	now := time.Now()
	datePrefix := now.Format("20060102")

	query := `
		SELECT COALESCE(MAX(CAST(SUBSTRING(order_no FROM '[0-9]+$') AS INTEGER)), 0) + 1
		FROM orders
		WHERE company_id = $1
		  AND branch_id = $2
		  AND order_no LIKE $3
	`

	pattern := fmt.Sprintf("ORD-%s-%%", datePrefix)
	var nextNum int
	if err := conn(ctx, r.db).GetContext(ctx, &nextNum, query, companyID, branchID, pattern); err != nil {
		nextNum = 1
	}

	return fmt.Sprintf("ORD-%s-%04d", datePrefix, nextNum), nil
}
