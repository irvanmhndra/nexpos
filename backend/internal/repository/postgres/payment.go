package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/jmoiron/sqlx"
)

type PaymentRepository struct {
	db *sqlx.DB
}

func NewPaymentRepository(db *sqlx.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

const paymentColumns = `id, order_id, method, amount, reference_no, status, refunded_amount, refunded_at, refund_reason, paid_at, created_at`

func (r *PaymentRepository) Create(ctx context.Context, payment *model.Payment) error {
	query := `
		INSERT INTO payments (order_id, method, amount, reference_no, status, paid_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`
	// Default status to completed for new payments
	if payment.Status == "" {
		payment.Status = model.PaymentStatusCompleted
	}
	return r.db.QueryRowContext(ctx, query,
		payment.OrderID,
		payment.Method,
		payment.Amount,
		payment.ReferenceNo,
		payment.Status,
		payment.PaidAt,
	).Scan(&payment.ID, &payment.CreatedAt)
}

func (r *PaymentRepository) GetByID(ctx context.Context, id int64) (*model.Payment, error) {
	var payment model.Payment
	query := fmt.Sprintf(`SELECT %s FROM payments WHERE id = $1`, paymentColumns)
	if err := r.db.GetContext(ctx, &payment, query, id); err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *PaymentRepository) GetByOrderID(ctx context.Context, orderID int64) ([]*model.Payment, error) {
	var payments []*model.Payment
	query := fmt.Sprintf(`SELECT %s FROM payments WHERE order_id = $1 ORDER BY id`, paymentColumns)
	if err := r.db.SelectContext(ctx, &payments, query, orderID); err != nil {
		return nil, err
	}
	return payments, nil
}

func (r *PaymentRepository) Update(ctx context.Context, payment *model.Payment) error {
	query := `
		UPDATE payments
		SET status = $1, refunded_amount = $2, refunded_at = $3, refund_reason = $4
		WHERE id = $5
	`
	result, err := r.db.ExecContext(ctx, query,
		payment.Status,
		payment.RefundedAmount,
		payment.RefundedAt,
		payment.RefundReason,
		payment.ID,
	)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("payment not found")
	}

	return nil
}

func (r *PaymentRepository) DeleteByOrderID(ctx context.Context, orderID int64) error {
	query := `DELETE FROM payments WHERE order_id = $1`
	_, err := r.db.ExecContext(ctx, query, orderID)
	return err
}

func (r *PaymentRepository) GetTotalPaidByOrderID(ctx context.Context, orderID int64) (float64, error) {
	var total float64
	query := `SELECT COALESCE(SUM(amount - refunded_amount), 0) FROM payments WHERE order_id = $1 AND status != 'failed'`
	if err := r.db.GetContext(ctx, &total, query, orderID); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *PaymentRepository) GetTotalRefundedByOrderID(ctx context.Context, orderID int64) (float64, error) {
	var total float64
	query := `SELECT COALESCE(SUM(refunded_amount), 0) FROM payments WHERE order_id = $1`
	if err := r.db.GetContext(ctx, &total, query, orderID); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *PaymentRepository) GetCashTotalByPeriod(ctx context.Context, branchID int64, from, to time.Time) (float64, error) {
	var total float64
	query := `
		SELECT COALESCE(SUM(p.amount - p.refunded_amount), 0)
		FROM payments p
		JOIN orders o ON o.id = p.order_id
		WHERE o.branch_id = $1
		  AND p.method = 'cash'
		  AND p.status != 'failed'
		  AND p.paid_at >= $2
		  AND p.paid_at < $3
	`
	if err := r.db.GetContext(ctx, &total, query, branchID, from, to); err != nil {
		return 0, err
	}
	return total, nil
}
